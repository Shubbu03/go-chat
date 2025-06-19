package websocket

import (
	"go-chat/internal/domain"

	"github.com/gorilla/websocket"
)

type WSHandler interface {
	HandleConnection(conn *websocket.Conn, userID uint) error
	DisconnectUser(userID uint)

	BroadcastMessage(message *domain.WSMessage, targetUserID uint) error
	BroadcastToAll(message *domain.WSMessage) error

	SetUserOnline(userID uint)
	SetUserOffline(userID uint)
	IsUserOnline(userID uint) bool
	GetOnlineUsers() []uint

	BroadcastTyping(senderID, receiverID uint) error
	BroadcastStopTyping(senderID, receiverID uint) error
}

type WSClient struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan *domain.WSMessage
}
