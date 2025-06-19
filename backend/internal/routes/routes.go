package routes

import (
	"go-chat/internal/handlers"
	"go-chat/internal/middlerware"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func SetupRoutes(r chi.Router, db *gorm.DB, h *handlers.Handlers) error {

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cookie"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Chat API"))
	})
	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", h.User.SignupHandler)
		r.Post("/login", h.User.LoginHandler)
		r.Post("/refresh", h.Auth.RefreshTokenHandler)
		r.Post("/logout", h.Auth.LogoutHandler)
		r.Post("/check-email", h.Auth.CheckEmailHandler)

		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.Get("/me", h.Auth.GetMeHandler)
			r.Get("/validate", h.Auth.ValidateTokenHandler)
			r.Post("/change-password", h.Auth.ChangePasswordHandler)
		})
	})

	// User routes
	r.Route("/api/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.Put("/profile", h.User.UpdateProfileHandler)
			r.Get("/search", h.User.SearchUsersHandler)
		})
	})

	// Friends routes
	r.Route("/api/friends", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)

			r.Get("/", h.Friends.GetUserFriendsHandler)
			r.Post("/request", h.Friends.SendFriendRequestHandler)
			r.Post("/{friendshipID}/accept", h.Friends.AcceptFriendRequestHandler)
			r.Post("/{friendshipID}/reject", h.Friends.RejectFriendRequestHandler)
			r.Delete("/{friendshipID}", h.Friends.RemoveFriendHandler)
			r.Post("/block", h.Friends.BlockUserHandler)
			r.Get("/requests/received", h.Friends.GetPendingFriendRequestsHandler)
			r.Get("/requests/sent", h.Friends.GetSentFriendRequestsHandler)
		})
	})

	// Message routes
	r.Route("/api/messages", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.Post("/", h.Message.SendMessageHandler)
			r.Get("/conversations", h.Message.GetConversationsHandler)
			r.Get("/conversations/search", h.Message.SearchConversationsHandler)
			r.Get("/{userID}", h.Message.GetMessagesHandler)
			r.Put("/read/{userID}", h.Message.MarkAsReadHandler)
			r.Get("/unread/{userID}", h.Message.GetUnreadCountHandler)
			r.Delete("/{messageID}", h.Message.DeleteMessageHandler)
			r.Get("/search", h.Message.SearchMessagesHandler)
		})
	})

	r.Get("/ws", h.WebSocket.ServeWS)

	return nil
}
