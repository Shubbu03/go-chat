package service

import (
	"errors"
	"time"

	"go-chat/internal/domain"
	"go-chat/internal/ports/repository"
)

type MessageService struct {
	repo     repository.MessageRepository
	userRepo repository.UserRepository
}

func NewMessageService(repo repository.MessageRepository, userRepo repository.UserRepository) *MessageService {
	return &MessageService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *MessageService) SendMessage(senderID uint, req *domain.MessageRequest) (*domain.MessageResponse, error) {
	_, err := s.userRepo.GetUserByID(req.ReceiverID)
	if err != nil {
		return nil, errors.New("receiver not found")
	}

	sender, err := s.userRepo.GetUserByID(senderID)
	if err != nil {
		return nil, errors.New("sender not found")
	}

	message := &domain.Message{
		SenderID:    senderID,
		ReceiverID:  req.ReceiverID,
		Content:     req.Content,
		MessageType: req.MessageType,
		IsRead:      false,
		IsDelivered: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if message.MessageType == "" {
		message.MessageType = "text"
	}

	err = s.repo.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	response := &domain.MessageResponse{
		ID:             message.ID,
		SenderID:       message.SenderID,
		ReceiverID:     message.ReceiverID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		IsRead:         message.IsRead,
		IsDelivered:    message.IsDelivered,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,
		SenderName:     sender.Name,
		SenderUsername: sender.Email,
	}

	return response, nil
}

func (s *MessageService) GetMessagesBetweenUsers(userID1, userID2 uint, limit, offset int) ([]*domain.MessageResponse, error) {
	_, err := s.userRepo.GetUserByID(userID1)
	if err != nil {
		return nil, errors.New("user1 not found")
	}

	_, err = s.userRepo.GetUserByID(userID2)
	if err != nil {
		return nil, errors.New("user2 not found")
	}

	messages, err := s.repo.GetMessagesBetweenUsers(userID1, userID2, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []*domain.MessageResponse
	for _, msg := range messages {
		response := &domain.MessageResponse{
			ID:             msg.ID,
			SenderID:       msg.SenderID,
			ReceiverID:     msg.ReceiverID,
			Content:        msg.Content,
			MessageType:    msg.MessageType,
			IsRead:         msg.IsRead,
			IsDelivered:    msg.IsDelivered,
			CreatedAt:      msg.CreatedAt,
			UpdatedAt:      msg.UpdatedAt,
			SenderName:     msg.Sender.Name,
			SenderUsername: msg.Sender.Email,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *MessageService) GetUserConversations(userID uint) ([]*domain.ConversationResponse, error) {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return s.repo.GetUserConversations(userID)
}

func (s *MessageService) MarkMessagesAsRead(senderID, receiverID uint) error {
	return s.repo.MarkMessagesAsRead(senderID, receiverID)
}

func (s *MessageService) MarkMessageAsDelivered(messageID uint) error {
	return s.repo.MarkMessageAsDelivered(messageID)
}

func (s *MessageService) GetUnreadMessageCount(senderID, receiverID uint) (int, error) {
	return s.repo.GetUnreadMessageCount(senderID, receiverID)
}

func (s *MessageService) DeleteMessage(messageID, userID uint) error {
	message, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return errors.New("message not found")
	}

	if message.SenderID != userID {
		return errors.New("unauthorized: can only delete your own messages")
	}

	return s.repo.DeleteMessage(messageID)
}

func (s *MessageService) SearchMessages(userID uint, query string, limit, offset int) ([]*domain.MessageResponse, error) {
	messages, err := s.repo.SearchMessages(userID, query, limit, offset)
	if err != nil {
		return nil, err
	}

	var responses []*domain.MessageResponse
	for _, msg := range messages {
		response := &domain.MessageResponse{
			ID:             msg.ID,
			SenderID:       msg.SenderID,
			ReceiverID:     msg.ReceiverID,
			Content:        msg.Content,
			MessageType:    msg.MessageType,
			IsRead:         msg.IsRead,
			IsDelivered:    msg.IsDelivered,
			CreatedAt:      msg.CreatedAt,
			UpdatedAt:      msg.UpdatedAt,
			SenderName:     msg.Sender.Name,
			SenderUsername: msg.Sender.Email,
		}
		responses = append(responses, response)
	}

	return responses, nil
}
