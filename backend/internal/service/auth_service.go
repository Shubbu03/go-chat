package service

import (
	"errors"
	"go-chat/internal/domain"
	"go-chat/internal/ports/repository"
	"go-chat/pkg"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) AuthenticateUser(email, password string) (*domain.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !pkg.ComparePassword(password, []byte(user.Password)) {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *AuthService) GenerateTokens(user *domain.User) (*pkg.TokenPair, error) {
	return pkg.GenerateTokenPair(user.ID, user.Email, user.Name)
}

func (s *AuthService) RefreshTokens(refreshToken string) (*pkg.TokenPair, *domain.User, error) {
	refreshClaims, err := pkg.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.userRepo.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}

	tokens, err := s.GenerateTokens(user)
	if err != nil {
		return nil, nil, errors.New("failed to generate tokens")
	}

	return tokens, user, nil
}

func (s *AuthService) ValidateAccessToken(token string) (*pkg.Claims, error) {
	return pkg.ValidateAccessToken(token)
}

func (s *AuthService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !pkg.ComparePassword(currentPassword, []byte(user.Password)) {
		return errors.New("current password is incorrect")
	}

	hashedNewPassword := pkg.HashPassword(newPassword)
	_, err = s.userRepo.UpdatePassword(userID, string(hashedNewPassword))
	if err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AuthService) CheckEmailExists(email string) (bool, error) {
	_, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *AuthService) GetUserByID(userID uint) (*domain.User, error) {
	return s.userRepo.GetUserByID(userID)
}

func (s *AuthService) ValidateTokenAndGetUser(token string) (*domain.User, error) {
	claims, err := s.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
