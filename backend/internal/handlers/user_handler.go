package handlers

import (
	"encoding/json"
	"go-chat/internal/domain"
	"go-chat/internal/middlerware"
	"go-chat/internal/service"
	"go-chat/pkg"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

func NewUserHandler(us *service.UserService, as *service.AuthService) *UserHandler {
	return &UserHandler{
		userService: us,
		authService: as,
	}
}

func (h *UserHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if email already exists
	exists, err := h.authService.CheckEmailExists(req.Email)
	if err != nil {
		http.Error(w, "Failed to check email availability", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	hashedPassword := pkg.HashPassword(req.Password)
	err = h.userService.Signup(req.Name, req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully",
	})
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	tokens, err := h.authService.GenerateTokens(user)
	if err != nil {
		log.Printf("Error generating tokens: %v", err)
		http.Error(w, "Could not generate authentication tokens", http.StatusInternalServerError)
		return
	}

	pkg.SetTokenCookies(w, tokens)

	log.Printf("✅ Email login successful for user: %s (ID: %d)", user.Email, user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Logged in successfully",
		"user":          user.ToResponse(),
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
	})
}

func (h *UserHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedUser, err := h.userService.UpdateProfile(userID, req.Name, req.Email)
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile updated successfully",
		"user":    updatedUser.ToResponse(),
	})
}

func (h *UserHandler) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	users, err := h.userService.SearchUsers(query, userID)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Convert to safe user responses
	userResponses := make([]*domain.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userResponses,
		"query": query,
		"count": len(userResponses),
	})
}
