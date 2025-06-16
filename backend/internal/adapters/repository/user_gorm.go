package repository_adapters

import (
	"go-chat/internal/domain"

	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
}

func toUserModel(u *domain.User) *UserModel {
	return &UserModel{
		Model:    gorm.Model{ID: u.ID},
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
	}
}

func toDomainUser(m *UserModel) *domain.User {
	return &domain.User{
		ID:       m.ID,
		Name:     m.Name,
		Email:    m.Email,
		Password: m.Password,
	}
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user *domain.User) error {
	return r.db.Create(toUserModel(user)).Error
}

func (r *GormUserRepository) FindByID(id uint) (*domain.User, error) {
	var model UserModel
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}
	return toDomainUser(&model), nil
}

func (r *GormUserRepository) FindByEmail(email string) (*domain.User, error) {
	var model UserModel
	if err := r.db.Where("email = ?", email).First(&model).Error; err != nil {
		return nil, err
	}
	return toDomainUser(&model), nil
}
