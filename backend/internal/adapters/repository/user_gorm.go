package repository_adapters

import (
	"go-chat/internal/domain"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func NewUserGormRepo(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *GormUserRepository) GetUserByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) UpdatePassword(id uint, password string) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	user.Password = password
	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) UpdateUserProfile(id uint, name, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}

	user.Name = name
	user.Email = email

	if err := r.db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) SearchUsers(query string, id uint) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.Where("name ILIKE ? AND id <> ?", "%"+query+"%", id).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}
