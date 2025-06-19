package repository_adapters

import (
	"go-chat/internal/domain"

	"gorm.io/gorm"
)

type FriendshipModel struct {
	gorm.Model
	RequesterID uint                      `gorm:"not null;index"`
	AddresseeID uint                      `gorm:"not null;index"`
	Status      domain.FriendshipStatus   `gorm:"not null;default:'pending'"`
	
	Requester UserModel `gorm:"foreignKey:RequesterID"`
	Addressee UserModel `gorm:"foreignKey:AddresseeID"`
}

func toFriendshipModel(f *domain.Friendship) *FriendshipModel {
	return &FriendshipModel{
		Model:       gorm.Model{ID: f.ID},
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      f.Status,
	}
}

func toDomainFriendship(f *FriendshipModel) *domain.Friendship {
	friendship := &domain.Friendship{
		ID:          f.ID,
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      f.Status,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
	
	if f.Requester.ID != 0 {
		friendship.Requester = toDomainUser(&f.Requester)
	}
	if f.Addressee.ID != 0 {
		friendship.Addressee = toDomainUser(&f.Addressee)
	}
	
	return friendship
}

type GormFriendsRepository struct {
	db *gorm.DB
}

func NewFriendsRepository(db *gorm.DB) *GormFriendsRepository {
	return &GormFriendsRepository{db: db}
}

func NewFriendsGormRepo(db *gorm.DB) *GormFriendsRepository {
	return &GormFriendsRepository{db: db}
}

func (r *GormFriendsRepository) CreateFriendship(friendship *domain.Friendship) error {
	model := toFriendshipModel(friendship)
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	friendship.ID = model.ID
	return nil
}

func (r *GormFriendsRepository) FindFriendshipByID(id uint) (*domain.Friendship, error) {
	var model FriendshipModel
	if err := r.db.Preload("Requester").Preload("Addressee").First(&model, id).Error; err != nil {
		return nil, err
	}
	return toDomainFriendship(&model), nil
}

func (r *GormFriendsRepository) FindFriendshipBetweenUsers(userID1, userID2 uint) (*domain.Friendship, error) {
	var model FriendshipModel
	if err := r.db.Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		userID1, userID2, userID2, userID1,
	).Preload("Requester").Preload("Addressee").First(&model).Error; err != nil {
		return nil, err
	}
	return toDomainFriendship(&model), nil
}

func (r *GormFriendsRepository) UpdateFriendshipStatus(id uint, status domain.FriendshipStatus) error {
	return r.db.Model(&FriendshipModel{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GormFriendsRepository) GetUserFriends(userID uint) ([]*domain.Friendship, error) {
	var models []FriendshipModel
	if err := r.db.Where(
		"(requester_id = ? OR addressee_id = ?) AND status = ?",
		userID, userID, domain.FriendshipAccepted,
	).Preload("Requester").Preload("Addressee").Find(&models).Error; err != nil {
		return nil, err
	}
	
	friendships := make([]*domain.Friendship, len(models))
	for i, model := range models {
		friendships[i] = toDomainFriendship(&model)
	}
	return friendships, nil
}

func (r *GormFriendsRepository) GetPendingFriendRequests(userID uint) ([]*domain.Friendship, error) {
	var models []FriendshipModel
	if err := r.db.Where(
		"addressee_id = ? AND status = ?",
		userID, domain.FriendshipPending,
	).Preload("Requester").Preload("Addressee").Find(&models).Error; err != nil {
		return nil, err
	}
	
	friendships := make([]*domain.Friendship, len(models))
	for i, model := range models {
		friendships[i] = toDomainFriendship(&model)
	}
	return friendships, nil
}

func (r *GormFriendsRepository) GetSentFriendRequests(userID uint) ([]*domain.Friendship, error) {
	var models []FriendshipModel
	if err := r.db.Where(
		"requester_id = ? AND status = ?",
		userID, domain.FriendshipPending,
	).Preload("Requester").Preload("Addressee").Find(&models).Error; err != nil {
		return nil, err
	}
	
	friendships := make([]*domain.Friendship, len(models))
	for i, model := range models {
		friendships[i] = toDomainFriendship(&model)
	}
	return friendships, nil
}

func (r *GormFriendsRepository) DeleteFriendship(id uint) error {
	return r.db.Delete(&FriendshipModel{}, id).Error
}

func (r *GormFriendsRepository) AreFriends(userID1, userID2 uint) (bool, error) {
	var count int64
	err := r.db.Model(&FriendshipModel{}).Where(
		"((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)) AND status = ?",
		userID1, userID2, userID2, userID1, domain.FriendshipAccepted,
	).Count(&count).Error
	
	return count > 0, err
}
