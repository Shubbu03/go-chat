package middlerware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-chat/pkg"
)

type RateLimiter struct {
	ipLimiters map[string]*TokenBucket
	ipMutex    sync.RWMutex

	userLimiters map[uint]*TokenBucket
	userMutex    sync.RWMutex

	cleanupTicker *time.Ticker
}

type TokenBucket struct {
	capacity     int
	tokens       int
	refillRate   int
	lastRefill   time.Time
	mutex        sync.Mutex
}

type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

var rateLimitConfigs = map[string]RateLimitConfig{
	"auth": {
		RequestsPerMinute: 10,
		BurstSize:         5,
	},
	"message": {
		RequestsPerMinute: 60,
		BurstSize:         10,
	},
	"search": {
		RequestsPerMinute: 30,
		BurstSize:         5,
	},
	"friends": {
		RequestsPerMinute: 20
		BurstSize:         5,
	},
	"default": {
		RequestsPerMinute: 100,
		BurstSize:         20,
	},
}

var globalRateLimiter *RateLimiter

func InitRateLimiter() {
	globalRateLimiter = &RateLimiter{
		ipLimiters:   make(map[string]*TokenBucket),
		userLimiters: make(map[uint]*TokenBucket),
	}

	globalRateLimiter.cleanupTicker = time.NewTicker(10 * time.Minute)
	go globalRateLimiter.cleanup()
}

func NewTokenBucket(config RateLimitConfig) *TokenBucket {
	return &TokenBucket{
		capacity:   config.BurstSize,
		tokens:     config.BurstSize,
		refillRate: config.RequestsPerMinute / 60,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tokensToAdd := int(elapsed * float64(tb.refillRate))
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func RateLimit(endpointType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if globalRateLimiter == nil {
				InitRateLimiter()
			}

			config, exists := rateLimitConfigs[endpointType]
			if !exists {
				config = rateLimitConfigs["default"]
			}

			clientIP := getClientIP(r)

			if !globalRateLimiter.checkIPRateLimit(clientIP, config) {
				pkg.WriteErrorResponse(w, http.StatusTooManyRequests, 
					fmt.Sprintf("Rate limit exceeded for IP. Try again in a minute."))
				return
			}

			if userID := getUserIDFromContext(r); userID != 0 {
				if !globalRateLimiter.checkUserRateLimit(userID, config) {
					pkg.WriteErrorResponse(w, http.StatusTooManyRequests, 
						fmt.Sprintf("Rate limit exceeded for user. Try again in a minute."))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) checkIPRateLimit(ip string, config RateLimitConfig) bool {
	rl.ipMutex.RLock()
	limiter, exists := rl.ipLimiters[ip]
	rl.ipMutex.RUnlock()

	if !exists {
		rl.ipMutex.Lock()
		if limiter, exists = rl.ipLimiters[ip]; !exists {
			limiter = NewTokenBucket(config)
			rl.ipLimiters[ip] = limiter
		}
		rl.ipMutex.Unlock()
	}

	return limiter.Allow()
}

func (rl *RateLimiter) checkUserRateLimit(userID uint, config RateLimitConfig) bool {
	rl.userMutex.RLock()
	limiter, exists := rl.userLimiters[userID]
	rl.userMutex.RUnlock()

	if !exists {
		rl.userMutex.Lock()
		if limiter, exists = rl.userLimiters[userID]; !exists {
			limiter = NewTokenBucket(config)
			rl.userLimiters[userID] = limiter
		}
		rl.userMutex.Unlock()
	}

	return limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTicker.C {
		now := time.Now()
		cutoff := now.Add(-30 * time.Minute)

		rl.ipMutex.Lock()
		for ip, limiter := range rl.ipLimiters {
			limiter.mutex.Lock()
			if limiter.lastRefill.Before(cutoff) {
				delete(rl.ipLimiters, ip)
			}
			limiter.mutex.Unlock()
		}
		rl.ipMutex.Unlock()

		rl.userMutex.Lock()
		for userID, limiter := range rl.userLimiters {
			limiter.mutex.Lock()
			if limiter.lastRefill.Before(cutoff) {
				delete(rl.userLimiters, userID)
			}
			limiter.mutex.Unlock()
		}
		rl.userMutex.Unlock()
	}
}

func getClientIP(r *http.Request) string {
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		if idx := len(xForwardedFor); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range xForwardedFor {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xForwardedFor[:commaIdx]
				}
			}
			return xForwardedFor
		}
	}

	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	if r.RemoteAddr != "" {
		if colonIdx := len(r.RemoteAddr) - 1; colonIdx >= 0 {
			for i := colonIdx; i >= 0; i-- {
				if r.RemoteAddr[i] == ':' {
					return r.RemoteAddr[:i]
				}
			}
		}
		return r.RemoteAddr
	}

	return "unknown"
}

func getUserIDFromContext(r *http.Request) uint {
	if userID := r.Context().Value("userID"); userID != nil {
		if id, ok := userID.(uint); ok {
			return id
		}
		if idStr, ok := userID.(string); ok {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				return uint(id)
			}
		}
	}
	return 0
}

func StopRateLimiter() {
	if globalRateLimiter != nil && globalRateLimiter.cleanupTicker != nil {
		globalRateLimiter.cleanupTicker.Stop()
	}
} 