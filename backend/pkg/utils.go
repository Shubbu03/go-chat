package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

func HashPassword(passwordToHash string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordToHash), bcrypt.DefaultCost)

	if err != nil {
		fmt.Printf("Error occured while hashing password! %s", err)
	}

	return hashedPassword
}

func ComparePassword(passwordToCompare string, dbPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordToCompare), dbPassword)

	if err != nil {
		fmt.Printf("Passwords not same! %s", err)
		return false
	}

	return true
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func ExtractUserIDFromRequest(r *http.Request) (uint, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString != authHeader {
			return extractUserIDFromToken(tokenString)
		}
	}

	token := r.URL.Query().Get("token")
	if token != "" {
		return extractUserIDFromToken(token)
	}

	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return extractUserIDFromToken(cookie.Value)
	}

	return 0, jwt.ErrTokenNotValidYet
}

func extractUserIDFromToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDStr, ok := claims["user_id"].(string); ok {
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				return 0, err
			}
			return uint(userID), nil
		}
	}

	return 0, jwt.ErrTokenInvalidClaims
}
