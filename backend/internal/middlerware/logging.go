package middlerware

import (
	"net/http"
	"strconv"
	"time"

	"go-chat/pkg"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(statusCode int) {
	if !lrw.written {
		lrw.statusCode = statusCode
		lrw.written = true
		lrw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (lrw *LoggingResponseWriter) Write(data []byte) (int, error) {
	if !lrw.written {
		lrw.WriteHeader(http.StatusOK)
	}
	return lrw.ResponseWriter.Write(data)
}

func RequestLogging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := NewLoggingResponseWriter(w)

			var userID uint
			if uid := r.Context().Value("userID"); uid != nil {
				if id, ok := uid.(uint); ok {
					userID = id
				}
			}

			clientIP := getClientIP(r)

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			if lrw.statusCode >= 400 {
				var err error
				if lrw.statusCode >= 500 {
					err = &HTTPError{StatusCode: lrw.statusCode, Message: "Server error"}
				} else {
					err = &HTTPError{StatusCode: lrw.statusCode, Message: "Client error"}
				}

				pkg.LogHTTPError(r.Method, r.URL.Path, clientIP, lrw.statusCode, err, userID)
			} else {
				pkg.LogHTTPRequest(r.Method, r.URL.Path, r.UserAgent(), clientIP, userID, duration)
			}

			if duration > time.Second {
				pkg.Warn("Slow HTTP request", map[string]interface{}{
					"method":    r.Method,
					"path":      r.URL.Path,
					"duration":  duration.String(),
					"client_ip": clientIP,
					"user_id":   userID,
				})
			}
		})
	}
}

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return e.Message + " (status: " + strconv.Itoa(e.StatusCode) + ")"
}

func AuthLogging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := NewLoggingResponseWriter(w)
			clientIP := getClientIP(r)

			email := ""
			if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
			}

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)
			success := lrw.statusCode < 400

			var event string
			switch r.URL.Path {
			case "/api/auth/login":
				event = "login"
			case "/api/auth/signup":
				event = "signup"
			case "/api/auth/logout":
				event = "logout"
			case "/api/auth/refresh":
				event = "refresh_token"
			default:
				event = "auth_request"
			}

			var userID uint
			if uid := r.Context().Value("userID"); uid != nil {
				if id, ok := uid.(uint); ok {
					userID = id
				}
			}

			pkg.LogAuthEvent(event, email, clientIP, userID, success)

			if success {
				pkg.LogHTTPRequest(r.Method, r.URL.Path, r.UserAgent(), clientIP, userID, duration)
			} else {
				err := &HTTPError{StatusCode: lrw.statusCode, Message: "Auth failed"}
				pkg.LogHTTPError(r.Method, r.URL.Path, clientIP, lrw.statusCode, err, userID)
			}
		})
	}
}

func LogDatabaseOperationWithDefer(operation, table string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		pkg.LogDatabaseOperation(operation, table, duration, nil)
	}
}

func RecoveryLogging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					clientIP := getClientIP(r)
					var userID uint
					if uid := r.Context().Value("userID"); uid != nil {
						if id, ok := uid.(uint); ok {
							userID = id
						}
					}

					pkg.Error("HTTP request panic", &PanicError{Value: rec}, map[string]interface{}{
						"method":    r.Method,
						"path":      r.URL.Path,
						"client_ip": clientIP,
						"user_id":   userID,
					})

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

type PanicError struct {
	Value interface{}
}

func (e *PanicError) Error() string {
	return "panic: " + pkg.ToString(e.Value)
}

func RateLimitLogging(endpointType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			var userID uint
			if uid := r.Context().Value("userID"); uid != nil {
				if id, ok := uid.(uint); ok {
					userID = id
				}
			}

			rateLimitMiddleware := RateLimit(endpointType)
			
			var rateLimitExceeded bool
			wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})

			rateLimitWriter := &rateLimitResponseWriter{
				ResponseWriter: w,
				onTooManyRequests: func() {
					rateLimitExceeded = true
				},
			}

			rateLimitMiddleware(wrappedHandler).ServeHTTP(rateLimitWriter, r)
			pkg.LogRateLimitEvent(clientIP, userID, endpointType, rateLimitExceeded)
		})
	}
}

type rateLimitResponseWriter struct {
	http.ResponseWriter
	onTooManyRequests func()
}

func (w *rateLimitResponseWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusTooManyRequests {
		w.onTooManyRequests()
	}
	w.ResponseWriter.WriteHeader(statusCode)
} 