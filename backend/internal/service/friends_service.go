package service

import (
	"errors"
	"go-chat/internal/domain"
	"go-chat/internal/ports/repository"
)

type FriendsService struct {
	repo repository.FriendsRepository
}

func NewFriendsService(repo repository.FriendsRepository) *FriendsService {
	return &FriendsService{repo: repo}
}

func (s *FriendsService) SendFriendRequest(requesterID, addresseeID uint) error {
	if requesterID == addresseeID {
		return errors.New("cannot send friend request to yourself")
	}

	existing, err := s.repo.FindFriendshipBetweenUsers(requesterID, addresseeID)
	if err == nil && existing != nil {
		switch existing.Status {
		case domain.FriendshipAccepted:
			return errors.New("users are already friends")
		case domain.FriendshipPending:
			return errors.New("friend request already pending")
		case domain.FriendshipBlocked:
			return errors.New("cannot send friend request to blocked user")
		}
	}

	friendship := &domain.Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      domain.FriendshipPending,
	}

	return s.repo.CreateFriendship(friendship)
}

func (s *FriendsService) AcceptFriendRequest(friendshipID, userID uint) error {
	friendship, err := s.repo.FindFriendshipByID(friendshipID)
	if err != nil {
		return err
	}

	if friendship.AddresseeID != userID {
		return errors.New("unauthorized to accept this friend request")
	}

	if friendship.Status != domain.FriendshipPending {
		return errors.New("friend request is not pending")
	}

	return s.repo.UpdateFriendshipStatus(friendshipID, domain.FriendshipAccepted)
}

func (s *FriendsService) RejectFriendRequest(friendshipID, userID uint) error {
	friendship, err := s.repo.FindFriendshipByID(friendshipID)
	if err != nil {
		return err
	}

	if friendship.AddresseeID != userID {
		return errors.New("unauthorized to reject this friend request")
	}

	if friendship.Status != domain.FriendshipPending {
		return errors.New("friend request is not pending")
	}

	return s.repo.UpdateFriendshipStatus(friendshipID, domain.FriendshipRejected)
}

func (s *FriendsService) RemoveFriend(friendshipID, userID uint) error {
	friendship, err := s.repo.FindFriendshipByID(friendshipID)
	if err != nil {
		return err
	}

	if friendship.RequesterID != userID && friendship.AddresseeID != userID {
		return errors.New("unauthorized to remove this friendship")
	}

	return s.repo.DeleteFriendship(friendshipID)
}

func (s *FriendsService) BlockUser(blockerID, blockedID uint) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	existing, err := s.repo.FindFriendshipBetweenUsers(blockerID, blockedID)
	if err != nil || existing == nil {
		friendship := &domain.Friendship{
			RequesterID: blockerID,
			AddresseeID: blockedID,
			Status:      domain.FriendshipBlocked,
		}
		return s.repo.CreateFriendship(friendship)
	}

	return s.repo.UpdateFriendshipStatus(existing.ID, domain.FriendshipBlocked)
}

func (s *FriendsService) GetUserFriends(userID uint) ([]*domain.Friendship, error) {
	return s.repo.GetUserFriends(userID)
}

func (s *FriendsService) GetPendingFriendRequests(userID uint) ([]*domain.Friendship, error) {
	return s.repo.GetPendingFriendRequests(userID)
}

func (s *FriendsService) GetSentFriendRequests(userID uint) ([]*domain.Friendship, error) {
	return s.repo.GetSentFriendRequests(userID)
}

func (s *FriendsService) AreFriends(userID1, userID2 uint) (bool, error) {
	return s.repo.AreFriends(userID1, userID2)
}
