package websocket_adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go-chat/internal/domain"
	wsports "go-chat/internal/ports/websocket"
	"go-chat/internal/service"
	"go-chat/pkg"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSHub struct {
	clients map[uint]*wsports.WSClient

	register chan *wsports.WSClient

	unregister chan *wsports.WSClient

	broadcast chan *domain.WSMessage

	mu sync.RWMutex

	messageService *service.MessageService
}

// Ensure WSHub implements WSHandler interface
var _ wsports.WSHandler = (*WSHub)(nil)

func NewWSHub(messageService *service.MessageService) *WSHub {
	return &WSHub{
		clients:        make(map[uint]*wsports.WSClient),
		register:       make(chan *wsports.WSClient),
		unregister:     make(chan *wsports.WSClient),
		broadcast:      make(chan *domain.WSMessage),
		messageService: messageService,
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()

			log.Printf("User %d connected via WebSocket", client.UserID)

			h.BroadcastToAll(&domain.WSMessage{
				Type: domain.WSMessageTypeUserOnline,
				Payload: map[string]interface{}{
					"user_id": client.UserID,
				},
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()

			log.Printf("User %d disconnected from WebSocket", client.UserID)

			h.BroadcastToAll(&domain.WSMessage{
				Type: domain.WSMessageTypeUserOffline,
				Payload: map[string]interface{}{
					"user_id": client.UserID,
				},
			})

		case message := <-h.broadcast:
			h.handleMessage(message)
		}
	}
}

func (h *WSHub) handleMessage(wsMsg *domain.WSMessage) {
	switch wsMsg.Type {
	case domain.WSMessageTypeNewMessage:
		if payload, ok := wsMsg.Payload.(map[string]interface{}); ok {
			if receiverID, ok := payload["receiver_id"].(float64); ok {
				h.BroadcastMessage(wsMsg, uint(receiverID))
			}
		}
	case domain.WSMessageTypeTyping, domain.WSMessageTypeStopTyping:
		if payload, ok := wsMsg.Payload.(map[string]interface{}); ok {
			if receiverID, ok := payload["receiver_id"].(float64); ok {
				h.BroadcastMessage(wsMsg, uint(receiverID))
			}
		}
	default:
		h.BroadcastToAll(wsMsg)
	}
}

func (h *WSHub) HandleConnection(conn *websocket.Conn, userID uint) error {
	client := &wsports.WSClient{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan *domain.WSMessage, 256),
	}

	h.register <- client

	go h.readPump(client)
	go h.writePump(client)

	return nil
}

func (h *WSHub) readPump(client *wsports.WSClient) {
	defer func() {
		h.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg domain.WSMessage
		if err := json.Unmarshal(messageBytes, &wsMsg); err != nil {
			log.Printf("Error unmarshaling WebSocket message: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "send_message":
			h.handleSendMessage(client, &wsMsg)
		case domain.WSMessageTypeTyping:
			h.handleTyping(client, &wsMsg)
		case domain.WSMessageTypeStopTyping:
			h.handleStopTyping(client, &wsMsg)
		case "mark_read":
			h.handleMarkAsRead(client, &wsMsg)
		}
	}
}

func (h *WSHub) writePump(client *wsports.WSClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling WebSocket message: %v", err)
				continue
			}

			if err := client.Conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("Error writing WebSocket message: %v", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHub) handleSendMessage(client *wsports.WSClient, wsMsg *domain.WSMessage) {
	payload, ok := wsMsg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	receiverID, ok := payload["receiver_id"].(float64)
	if !ok {
		return
	}

	content, ok := payload["content"].(string)
	if !ok {
		return
	}

	messageType, _ := payload["message_type"].(string)
	if messageType == "" {
		messageType = "text"
	}

	req := &domain.MessageRequest{
		ReceiverID:  uint(receiverID),
		Content:     content,
		MessageType: messageType,
	}

	message, err := h.messageService.SendMessage(client.UserID, req)
	if err != nil {
		errorMsg := &domain.WSMessage{
			Type: "error",
			Payload: map[string]interface{}{
				"message": err.Error(),
			},
		}
		select {
		case client.Send <- errorMsg:
		default:
			close(client.Send)
		}
		return
	}

	if h.IsUserOnline(uint(receiverID)) {
		h.messageService.MarkMessageAsDelivered(message.ID)
	}

	wsPayload := &domain.WSMessagePayload{
		MessageID:      message.ID,
		SenderID:       message.SenderID,
		ReceiverID:     message.ReceiverID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		Timestamp:      message.CreatedAt,
		SenderName:     message.SenderName,
		SenderUsername: message.SenderUsername,
	}

	broadcastMsg := &domain.WSMessage{
		Type:    domain.WSMessageTypeNewMessage,
		Payload: wsPayload,
	}

	h.BroadcastMessage(broadcastMsg, uint(receiverID))

	confirmMsg := &domain.WSMessage{
		Type:    "message_sent",
		Payload: wsPayload,
	}
	select {
	case client.Send <- confirmMsg:
	default:
		close(client.Send)
	}
}

func (h *WSHub) handleTyping(client *wsports.WSClient, wsMsg *domain.WSMessage) {
	payload, ok := wsMsg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	receiverID, ok := payload["receiver_id"].(float64)
	if !ok {
		return
	}

	typingMsg := &domain.WSMessage{
		Type: domain.WSMessageTypeTyping,
		Payload: map[string]interface{}{
			"sender_id":   client.UserID,
			"receiver_id": uint(receiverID),
		},
	}

	h.BroadcastMessage(typingMsg, uint(receiverID))
}

func (h *WSHub) handleStopTyping(client *wsports.WSClient, wsMsg *domain.WSMessage) {
	payload, ok := wsMsg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	receiverID, ok := payload["receiver_id"].(float64)
	if !ok {
		return
	}

	stopTypingMsg := &domain.WSMessage{
		Type: domain.WSMessageTypeStopTyping,
		Payload: map[string]interface{}{
			"sender_id":   client.UserID,
			"receiver_id": uint(receiverID),
		},
	}

	h.BroadcastMessage(stopTypingMsg, uint(receiverID))
}

func (h *WSHub) handleMarkAsRead(client *wsports.WSClient, wsMsg *domain.WSMessage) {
	payload, ok := wsMsg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	senderID, ok := payload["sender_id"].(float64)
	if !ok {
		return
	}

	err := h.messageService.MarkMessagesAsRead(uint(senderID), client.UserID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		return
	}

	readMsg := &domain.WSMessage{
		Type: domain.WSMessageTypeMessageRead,
		Payload: map[string]interface{}{
			"sender_id":   uint(senderID),
			"receiver_id": client.UserID,
		},
	}

	h.BroadcastMessage(readMsg, uint(senderID))
}

func (h *WSHub) BroadcastMessage(message *domain.WSMessage, targetUserID uint) error {
	h.mu.RLock()
	client, exists := h.clients[targetUserID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("user %d is not connected", targetUserID)
	}

	select {
	case client.Send <- message:
		return nil
	default:
		close(client.Send)
		h.mu.Lock()
		delete(h.clients, targetUserID)
		h.mu.Unlock()
		return fmt.Errorf("failed to send message to user %d", targetUserID)
	}
}

func (h *WSHub) BroadcastToAll(message *domain.WSMessage) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for userID, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, userID)
		}
	}

	return nil
}

func (h *WSHub) DisconnectUser(userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, exists := h.clients[userID]; exists {
		close(client.Send)
		delete(h.clients, userID)
	}
}

func (h *WSHub) SetUserOnline(userID uint) {
}

func (h *WSHub) SetUserOffline(userID uint) {
	h.DisconnectUser(userID)
}

func (h *WSHub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.clients[userID]
	return exists
}

func (h *WSHub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uint, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}

	return users
}

func (h *WSHub) BroadcastTyping(senderID, receiverID uint) error {
	typingMsg := &domain.WSMessage{
		Type: domain.WSMessageTypeTyping,
		Payload: map[string]interface{}{
			"sender_id":   senderID,
			"receiver_id": receiverID,
		},
	}

	return h.BroadcastMessage(typingMsg, receiverID)
}

func (h *WSHub) BroadcastStopTyping(senderID, receiverID uint) error {
	stopTypingMsg := &domain.WSMessage{
		Type: domain.WSMessageTypeStopTyping,
		Payload: map[string]interface{}{
			"sender_id":   senderID,
			"receiver_id": receiverID,
		},
	}

	return h.BroadcastMessage(stopTypingMsg, receiverID)
}

func (h *WSHub) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, err := pkg.ExtractUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	err = h.HandleConnection(conn, userID)
	if err != nil {
		log.Printf("WebSocket connection error: %v", err)
		conn.Close()
	}
}
