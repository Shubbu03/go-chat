package pkg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

var (
	jwtSecret       []byte
	refreshSecret   []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("Warning: JWT_SECRET not set, using default (not secure for production)")
		secret = "your-super-secret-jwt-key-change-this-in-production"
	}
	jwtSecret = []byte(secret)

	refreshSecretStr := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecretStr == "" {
		log.Println("Warning: JWT_REFRESH_SECRET not set, using default (not secure for production)")
		refreshSecretStr = "your-super-secret-refresh-key-change-this-in-production"
	}
	refreshSecret = []byte(refreshSecretStr)

	accessTokenTTL = 15 * time.Minute
	if ttlStr := os.Getenv("JWT_ACCESS_TTL"); ttlStr != "" {
		if minutes, err := strconv.Atoi(ttlStr); err == nil {
			accessTokenTTL = time.Duration(minutes) * time.Minute
		}
	}

	refreshTokenTTL = 7 * 24 * time.Hour
	if ttlStr := os.Getenv("JWT_REFRESH_TTL"); ttlStr != "" {
		if hours, err := strconv.Atoi(ttlStr); err == nil {
			refreshTokenTTL = time.Duration(hours) * time.Hour
		}
	}

	log.Printf("✅ JWT service initialized - Access TTL: %v, Refresh TTL: %v", accessTokenTTL, refreshTokenTTL)
}

func GenerateTokenPair(userID uint, email, name string) (*TokenPair, error) {
	now := time.Now()

	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-chat-api",
			Subject:   fmt.Sprintf("user-%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("could not create access token: %w", err)
	}

	refreshClaims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-chat-api",
			Subject:   fmt.Sprintf("refresh-%d", userID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(refreshSecret)
	if err != nil {
		return nil, fmt.Errorf("could not create refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refreshSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

func ExtractUserIDFromRequest(r *http.Request) (uint, error) {
	tokenString, err := ExtractTokenFromRequest(r)
	if err != nil {
		return 0, err
	}

	claims, err := ValidateAccessToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

func ExtractTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if after, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
			return after, nil
		}
	}

	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value, nil
	}

	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no authentication token provided")
}

func ValidateTokenAndGetClaims(tokenString string) (*Claims, error) {
	return ValidateAccessToken(tokenString)
}

func GetAccessTokenTTL() time.Duration {
	return accessTokenTTL
}

func GetRefreshTokenTTL() time.Duration {
	return refreshTokenTTL
}
