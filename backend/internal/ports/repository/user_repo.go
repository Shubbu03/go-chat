package repository

import "go-chat/internal/domain"

type UserRepository interface {
	Create(user *domain.User) error
	GetUserByID(id uint) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	UpdatePassword(id uint, password string) (*domain.User, error)
	UpdateUserProfile(id uint, name, email string) (*domain.User, error)
	SearchUsers(query string, id uint) ([]*domain.User, error)
}
