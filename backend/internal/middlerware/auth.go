package middlerware

import (
	"context"
	"go-chat/pkg"
	"net/http"
	"strings"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			if after, ok :=strings.CutPrefix(authHeader, "Bearer "); ok  {
				tokenString = after
			}
		}
		
		if tokenString == "" {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, "No authentication token provided", http.StatusUnauthorized)
				return
			}
			tokenString = cookie.Value
		}
		
		claims, err := pkg.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		ctx = context.WithValue(ctx, "userEmail", claims.Email)
		ctx = context.WithValue(ctx, "userName", claims.Name)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value("userID").(uint)
	return userID, ok
}

func GetUserEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value("userEmail").(string)
	return email, ok
}

func GetUserNameFromContext(r *http.Request) (string, bool) {
	name, ok := r.Context().Value("userName").(string)
	return name, ok
}