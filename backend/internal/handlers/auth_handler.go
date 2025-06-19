package handlers

import (
	"encoding/json"
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

	refreshClaims, err := pkg.ValidateRefreshToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUserByID(refreshClaims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	tokens, err := pkg.GenerateTokenPair(user.ID, user.Email, user.Name)
	if err != nil {
		http.Error(w, "Could not generate new tokens", http.StatusInternalServerError)
		return
	}

	pkg.SetTokenCookies(w, tokens)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Token refreshed successfully",
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

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userResponse := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": userResponse,
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

	type changePasswordRequest struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if !pkg.ComparePassword(req.CurrentPassword, []byte(user.Password)) {
		http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		return
	}

	hashedNewPassword := pkg.HashPassword(req.NewPassword)

	_, err = h.userService.UpdatePasswordByID(userID, string(hashedNewPassword))
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}

func (h *AuthHandler) CheckEmailHandler(w http.ResponseWriter, r *http.Request) {
	type emailRequest struct {
		Email string `json:"email"`
	}

	var req emailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.userService.Login(req.Email)
	emailExists := err == nil

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": emailExists,
		"email":  req.Email,
	})
}
