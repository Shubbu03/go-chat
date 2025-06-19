package domain

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SenderID     uint      `json:"sender_id" gorm:"not null"`
	ReceiverID   uint      `json:"receiver_id" gorm:"not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	MessageType  string    `json:"message_type" gorm:"default:'text'"`
	IsRead       bool      `json:"is_read" gorm:"default:false"`
	IsDelivered  bool      `json:"is_delivered" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Sender   User `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Receiver User `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
}

type MessageRequest struct {
	ReceiverID  uint   `json:"receiver_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	MessageType string `json:"message_type,omitempty"`
}

type MessageResponse struct {
	ID          uint      `json:"id"`
	SenderID    uint      `json:"sender_id"`
	ReceiverID  uint      `json:"receiver_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	IsRead      bool      `json:"is_read"`
	IsDelivered bool      `json:"is_delivered"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	SenderName     string `json:"sender_name,omitempty"`
	SenderUsername string `json:"sender_username,omitempty"`
}

type ConversationResponse struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	LastMessage  *MessageResponse `json:"last_message,omitempty"`
	UnreadCount  int       `json:"unread_count"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WSMessagePayload struct {
	MessageID   uint   `json:"message_id"`
	SenderID    uint   `json:"sender_id"`
	ReceiverID  uint   `json:"receiver_id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
	Timestamp   time.Time `json:"timestamp"`
	SenderName  string `json:"sender_name"`
	SenderUsername string `json:"sender_username"`
}

const (
	WSMessageTypeNewMessage    = "new_message"
	WSMessageTypeMessageRead   = "message_read"
	WSMessageTypeTyping        = "typing"
	WSMessageTypeStopTyping    = "stop_typing"
	WSMessageTypeUserOnline    = "user_online"
	WSMessageTypeUserOffline   = "user_offline"
)