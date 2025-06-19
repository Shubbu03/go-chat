package domain

import "time"

type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipRejected FriendshipStatus = "rejected"
	FriendshipBlocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	ID          uint             `json:"id"`
	RequesterID uint             `json:"requester_id"`
	AddresseeID uint             `json:"addressee_id"`
	Status      FriendshipStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	
	Requester *User `json:"requester,omitempty"`
	Addressee *User `json:"addressee,omitempty"`
}

type FriendRequest struct {
	UserID uint `json:"user_id"`
}

type FriendResponse struct {
	FriendshipID uint             `json:"friendship_id"`
	Status       FriendshipStatus `json:"status"`
}
