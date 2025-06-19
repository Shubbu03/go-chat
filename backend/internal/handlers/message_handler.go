package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-chat/internal/domain"
	"go-chat/internal/service"
	"go-chat/pkg"

	"github.com/go-chi/chi/v5"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(ms *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: ms}
}

func (h *MessageHandler) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	
	var req domain.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.ReceiverID == 0 || req.Content == "" {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "receiver_id and content are required")
		return
	}
	
	if req.ReceiverID == userID {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Cannot send message to yourself")
		return
	}
	
	message, err := h.messageService.SendMessage(userID, &req)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Message sent successfully",
		"data":    message,
	})
}

func (h *MessageHandler) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("userID").(uint)
	
	otherUserIDStr := chi.URLParam(r, "userID")
	otherUserID, err := strconv.ParseUint(otherUserIDStr, 10, 32)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 50 
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	messages, err := h.messageService.GetMessagesBetweenUsers(currentUserID, uint(otherUserID), limit, offset)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": messages,
		"meta": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"count":  len(messages),
		},
	})
}

func (h *MessageHandler) GetConversationsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	
	conversations, err := h.messageService.GetUserConversations(userID)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": conversations,
	})
}

func (h *MessageHandler) MarkAsReadHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("userID").(uint)
	
	senderUserIDStr := chi.URLParam(r, "userID")
	senderUserID, err := strconv.ParseUint(senderUserIDStr, 10, 32)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	err = h.messageService.MarkMessagesAsRead(uint(senderUserID), currentUserID)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Messages marked as read",
	})
}

func (h *MessageHandler) GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("userID").(uint)
	
	senderUserIDStr := chi.URLParam(r, "userID")
	senderUserID, err := strconv.ParseUint(senderUserIDStr, 10, 32)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	
	count, err := h.messageService.GetUnreadMessageCount(uint(senderUserID), currentUserID)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"unread_count": count,
	})
}

func (h *MessageHandler) DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	
	messageIDStr := chi.URLParam(r, "messageID")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Invalid message ID")
		return
	}
	
	err = h.messageService.DeleteMessage(uint(messageID), userID)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Message deleted successfully",
	})
}

func (h *MessageHandler) SearchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(uint)
	
	query := r.URL.Query().Get("q")
	if query == "" {
		pkg.WriteErrorResponse(w, http.StatusBadRequest, "Search query is required")
		return
	}
	
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 20
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	messages, err := h.messageService.SearchMessages(userID, query, limit, offset)
	if err != nil {
		pkg.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	pkg.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": messages,
		"meta": map[string]interface{}{
			"query":  query,
			"limit":  limit,
			"offset": offset,
			"count":  len(messages),
		},
	})
}