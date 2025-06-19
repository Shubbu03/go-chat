package handlers

import (
	repository_adapters "go-chat/internal/adapters/repository"
	websocket_adapters "go-chat/internal/adapters/websocket"
	"go-chat/internal/service"

	"gorm.io/gorm"
)

type Handlers struct {
	Auth      *AuthHandler
	User      *UserHandler
	Friends   *FriendsHandler
	Message   *MessageHandler
	WebSocket *websocket_adapters.WSHub
}

func NewHandlers(db *gorm.DB) *Handlers {
	userRepo := repository_adapters.NewUserGormRepo(db)
	friendsRepo := repository_adapters.NewFriendsGormRepo(db)
	messageRepo := repository_adapters.NewMessageGormRepo(db)

	authService := service.NewAuthService()
	userService := service.NewUserService(userRepo)
	friendsService := service.NewFriendsService(friendsRepo)
	messageService := service.NewMessageService(messageRepo, userRepo)

	wsHub := websocket_adapters.NewWSHub(messageService)

	go wsHub.Run()

	authHandler := NewAuthHandler(authService, userService)
	userHandler := NewUserHandler(userService)
	friendsHandler := NewFriendsHandler(friendsService)
	messageHandler := NewMessageHandler(messageService)

	return &Handlers{
		Auth:      authHandler,
		User:      userHandler,
		Friends:   friendsHandler,
		Message:   messageHandler,
		WebSocket: wsHub,
	}
}
