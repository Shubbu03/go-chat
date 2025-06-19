package repository_adapters

import (
	"fmt"
	"time"

	"go-chat/internal/domain"
	"go-chat/internal/ports/repository"
	"gorm.io/gorm"
)

type messageGormRepo struct {
	db *gorm.DB
}

func NewMessageGormRepo(db *gorm.DB) repository.MessageRepository {
	return &messageGormRepo{db: db}
}

func (r *messageGormRepo) CreateMessage(message *domain.Message) error {
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()
	
	return r.db.Create(message).Error
}

func (r *messageGormRepo) GetMessagesBetweenUsers(userID1, userID2 uint, limit, offset int) ([]*domain.Message, error) {
	var messages []*domain.Message
	
	err := r.db.
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", 
			userID1, userID2, userID2, userID1).
		Preload("Sender").
		Preload("Receiver").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	
	return messages, err
}

func (r *messageGormRepo) GetUserConversations(userID uint) ([]*domain.ConversationResponse, error) {
	type ConversationUser struct {
		UserID      uint      `json:"user_id"`
		Username    string    `json:"username"`
		FullName    string    `json:"full_name"`
		LastMsgTime time.Time `json:"last_msg_time"`
	}
	
	var conversationUsers []ConversationUser
	
	err := r.db.Raw(`
		SELECT DISTINCT 
			CASE 
				WHEN m.sender_id = ? THEN m.receiver_id 
				ELSE m.sender_id 
			END as user_id,
			u.email as username,
			u.name as full_name,
			MAX(m.created_at) as last_msg_time
		FROM messages m
		JOIN users u ON u.id = CASE 
			WHEN m.sender_id = ? THEN m.receiver_id 
			ELSE m.sender_id 
		END
		WHERE (m.sender_id = ? OR m.receiver_id = ?) 
			AND m.deleted_at IS NULL 
			AND u.deleted_at IS NULL
		GROUP BY 
			CASE 
				WHEN m.sender_id = ? THEN m.receiver_id 
				ELSE m.sender_id 
			END, u.email, u.name
		ORDER BY last_msg_time DESC
	`, userID, userID, userID, userID, userID).Scan(&conversationUsers).Error
	
	if err != nil {
		return nil, err
	}
	
	var conversations []*domain.ConversationResponse
	
	for _, convUser := range conversationUsers {
		conv := &domain.ConversationResponse{
			UserID:   convUser.UserID,
			Username: convUser.Username,
			FullName: convUser.FullName,
		}
		
		lastMessage, err := r.GetLatestMessageBetweenUsers(userID, conv.UserID)
		if err == nil && lastMessage != nil {
			conv.LastMessage = &domain.MessageResponse{
				ID:             lastMessage.ID,
				SenderID:       lastMessage.SenderID,
				ReceiverID:     lastMessage.ReceiverID,
				Content:        lastMessage.Content,
				MessageType:    lastMessage.MessageType,
				IsRead:         lastMessage.IsRead,
				IsDelivered:    lastMessage.IsDelivered,
				CreatedAt:      lastMessage.CreatedAt,
				UpdatedAt:      lastMessage.UpdatedAt,
				SenderName:     lastMessage.Sender.Name,
				SenderUsername: lastMessage.Sender.Email,
			}
		}
		
		unreadCount, err := r.GetUnreadMessageCount(conv.UserID, userID)
		if err == nil {
			conv.UnreadCount = unreadCount
		}
		
		conversations = append(conversations, conv)
	}
	
	return conversations, nil
}

func (r *messageGormRepo) MarkMessagesAsRead(senderID, receiverID uint) error {
	return r.db.Model(&domain.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = false", senderID, receiverID).
		Update("is_read", true).Error
}

func (r *messageGormRepo) MarkMessageAsDelivered(messageID uint) error {
	return r.db.Model(&domain.Message{}).
		Where("id = ?", messageID).
		Update("is_delivered", true).Error
}

func (r *messageGormRepo) GetUnreadMessageCount(senderID, receiverID uint) (int, error) {
	var count int64
	err := r.db.Model(&domain.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = false", senderID, receiverID).
		Count(&count).Error
	
	return int(count), err
}

func (r *messageGormRepo) GetMessageByID(messageID uint) (*domain.Message, error) {
	var message domain.Message
	err := r.db.
		Preload("Sender").
		Preload("Receiver").
		First(&message, messageID).Error
	
	return &message, err
}

func (r *messageGormRepo) DeleteMessage(messageID uint) error {
	return r.db.Delete(&domain.Message{}, messageID).Error
}

func (r *messageGormRepo) GetLatestMessageBetweenUsers(userID1, userID2 uint) (*domain.Message, error) {
	var message domain.Message
	
	err := r.db.
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", 
			userID1, userID2, userID2, userID1).
		Preload("Sender").
		Preload("Receiver").
		Order("created_at DESC").
		First(&message).Error
	
	if err != nil {
		return nil, err
	}
	
	return &message, nil
}

func (r *messageGormRepo) SearchMessages(userID uint, query string, limit, offset int) ([]*domain.Message, error) {
	var messages []*domain.Message
	
	searchQuery := fmt.Sprintf("%%%s%%", query)
	
	err := r.db.
		Joins("LEFT JOIN users sender ON messages.sender_id = sender.id").
		Joins("LEFT JOIN users receiver ON messages.receiver_id = receiver.id").
		Where(`(messages.sender_id = ? OR messages.receiver_id = ?) AND 
			   (messages.content ILIKE ? OR 
			    sender.name ILIKE ? OR 
			    receiver.name ILIKE ?)`, 
			userID, userID, searchQuery, searchQuery, searchQuery).
		Preload("Sender").
		Preload("Receiver").
		Order("messages.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	
	return messages, err
}