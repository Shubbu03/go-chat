package handlers

import (
	"encoding/json"
	"go-chat/internal/middlerware"
	"go-chat/internal/service"
	"go-chat/pkg"
	"log"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	hashedPassword := pkg.HashPassword(body.Password)
	err := h.userService.Signup(body.Name, body.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created"))
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Login(body.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	validPassword := pkg.ComparePassword(body.Password, []byte(user.Password))
	if !validPassword {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	tokens, err := pkg.GenerateTokenPair(user.ID, user.Email, user.Name)
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

	type updateProfileRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	var req updateProfileRequest
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
		"user":    updatedUser,
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"query": query,
	})
}
