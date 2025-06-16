package service

import (
	"go-chat/internal/domain"
	"go-chat/internal/ports/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Signup(name, email, hashedPassword string) error {
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	}
	return s.repo.Create(user)
}

func (s *UserService) Login(email string) (*domain.User, error) {
	return s.repo.FindByEmail(email)
}

func (s *UserService) GetUserByID(id uint) (*domain.User, error) {
	return s.repo.FindByID(id)
}