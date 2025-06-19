package handlers

import (
	"encoding/json"
	"go-chat/internal/domain"
	"go-chat/internal/middlerware"
	"go-chat/internal/service"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type FriendsHandler struct {
	friendsService *service.FriendsService
}

func NewFriendsHandler(fs *service.FriendsService) *FriendsHandler {
	return &FriendsHandler{friendsService: fs}
}

func (h *FriendsHandler) SendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := middlerware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.friendsService.SendFriendRequest(userID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request sent successfully",
	})
}

func (h *FriendsHandler) AcceptFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friendshipIDStr := chi.URLParam(r, "friendshipID")
	friendshipID, err := strconv.ParseUint(friendshipIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid friendship ID", http.StatusBadRequest)
		return
	}

	if err := h.friendsService.AcceptFriendRequest(uint(friendshipID), userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request accepted",
	})
}

func (h *FriendsHandler) RejectFriendRequestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friendshipIDStr := chi.URLParam(r, "friendshipID")
	friendshipID, err := strconv.ParseUint(friendshipIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid friendship ID", http.StatusBadRequest)
		return
	}

	if err := h.friendsService.RejectFriendRequest(uint(friendshipID), userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request rejected",
	})
}

func (h *FriendsHandler) RemoveFriendHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friendshipIDStr := chi.URLParam(r, "friendshipID")
	friendshipID, err := strconv.ParseUint(friendshipIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid friendship ID", http.StatusBadRequest)
		return
	}

	if err := h.friendsService.RemoveFriend(uint(friendshipID), userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend removed successfully",
	})
}

func (h *FriendsHandler) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req domain.FriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.friendsService.BlockUser(userID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User blocked successfully",
	})
}

func (h *FriendsHandler) GetUserFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	friends, err := h.friendsService.GetUserFriends(userID)
	if err != nil {
		http.Error(w, "Failed to get friends", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"friends": friends,
		"count":   len(friends),
	})
}

func (h *FriendsHandler) GetPendingFriendRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := h.friendsService.GetPendingFriendRequests(userID)
	if err != nil {
		http.Error(w, "Failed to get friend requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"count":    len(requests),
	})
}

func (h *FriendsHandler) GetSentFriendRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requests, err := h.friendsService.GetSentFriendRequests(userID)
	if err != nil {
		http.Error(w, "Failed to get sent requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"count":    len(requests),
	})
}