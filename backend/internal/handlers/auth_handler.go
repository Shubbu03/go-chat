package handlers

import (
	"encoding/json"
	"go-chat/internal/domain"
	"go-chat/internal/middlerware"
	"go-chat/internal/service"
	"go-chat/pkg"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(as *service.AuthService, us *service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: as,
		userService: us,
	}
}

func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "No refresh token provided", http.StatusUnauthorized)
		return
	}

	tokens, user, err := h.authService.RefreshTokens(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	pkg.SetTokenCookies(w, tokens)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
		"user":          user.ToResponse(),
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	pkg.ClearTokenCookies(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user.ToResponse(),
	})
}

func (h *AuthHandler) ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"user_id": userID,
	})
}

func (h *AuthHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

func (h *AuthHandler) CheckEmailHandler(w http.ResponseWriter, r *http.Request) {
	type emailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req emailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	exists, err := h.authService.CheckEmailExists(req.Email)
	if err != nil {
		http.Error(w, "Failed to check email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": exists,
		"email":  req.Email,
	})
}
