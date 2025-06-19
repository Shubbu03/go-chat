package repository

import "go-chat/internal/domain"

type MessageRepository interface {
	CreateMessage(message *domain.Message) error
	
	GetMessagesBetweenUsers(userID1, userID2 uint, limit, offset int) ([]*domain.Message, error)
	
	GetUserConversations(userID uint) ([]*domain.ConversationResponse, error)
	
	MarkMessagesAsRead(senderID, receiverID uint) error
	
	MarkMessageAsDelivered(messageID uint) error
	
	GetUnreadMessageCount(senderID, receiverID uint) (int, error)
	
	GetMessageByID(messageID uint) (*domain.Message, error)
	
	DeleteMessage(messageID uint) error
	
	GetLatestMessageBetweenUsers(userID1, userID2 uint) (*domain.Message, error)
	
	SearchMessages(userID uint, query string, limit, offset int) ([]*domain.Message, error)
}
