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
	var conversations []*domain.ConversationResponse
	
	query := `
		SELECT DISTINCT 
			CASE 
				WHEN m.sender_id = ? THEN m.receiver_id 
				ELSE m.sender_id 
			END as user_id,
			u.username,
			u.full_name
		FROM messages m
		JOIN users u ON u.id = CASE 
			WHEN m.sender_id = ? THEN m.receiver_id 
			ELSE m.sender_id 
		END
		WHERE m.sender_id = ? OR m.receiver_id = ?
		ORDER BY user_id
	`
	
	rows, err := r.db.Raw(query, userID, userID, userID, userID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var conv domain.ConversationResponse
		if err := rows.Scan(&conv.UserID, &conv.Username, &conv.FullName); err != nil {
			continue
		}
		
		lastMessage, _ := r.GetLatestMessageBetweenUsers(userID, conv.UserID)
		if lastMessage != nil {
			conv.LastMessage = &domain.MessageResponse{
				ID:          lastMessage.ID,
				SenderID:    lastMessage.SenderID,
				ReceiverID:  lastMessage.ReceiverID,
				Content:     lastMessage.Content,
				MessageType: lastMessage.MessageType,
				IsRead:      lastMessage.IsRead,
				IsDelivered: lastMessage.IsDelivered,
				CreatedAt:   lastMessage.CreatedAt,
				UpdatedAt:   lastMessage.UpdatedAt,
			}
		}
		
		unreadCount, _ := r.GetUnreadMessageCount(conv.UserID, userID)
		conv.UnreadCount = unreadCount
		
		conversations = append(conversations, &conv)
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
		Where("(sender_id = ? OR receiver_id = ?) AND content ILIKE ?", userID, userID, searchQuery).
		Preload("Sender").
		Preload("Receiver").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	
	return messages, err
}