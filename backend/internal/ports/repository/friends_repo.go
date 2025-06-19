package repository

import "go-chat/internal/domain"

type FriendsRepository interface {
	CreateFriendship(friendship *domain.Friendship) error
	
	FindFriendshipByID(id uint) (*domain.Friendship, error)
	
	FindFriendshipBetweenUsers(userID1, userID2 uint) (*domain.Friendship, error)
	
	UpdateFriendshipStatus(id uint, status domain.FriendshipStatus) error
	
	GetUserFriends(userID uint) ([]*domain.Friendship, error)
	
	GetPendingFriendRequests(userID uint) ([]*domain.Friendship, error)
	
	GetSentFriendRequests(userID uint) ([]*domain.Friendship, error)
	
	DeleteFriendship(id uint) error
	
	AreFriends(userID1, userID2 uint) (bool, error)
}
