package routes

import (
	"go-chat/internal/handlers"
	"go-chat/internal/middlerware"
	"go-chat/pkg"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

func SetupRoutes(r chi.Router, db *gorm.DB, h *handlers.Handlers) error {
	pkg.InitLogger(pkg.INFO, true)
	middlerware.InitRateLimiter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cookie"},
		ExposedHeaders:   []string{"Link", "Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middlerware.SecurityHeaders)
	r.Use(middlerware.RecoveryLogging())
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middlerware.RequestLogging())
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Chat API"))
	})

	r.Route("/api/auth", func(r chi.Router) {
		r.Use(middlerware.RateLimit("auth"))
		r.Use(middlerware.AuthLogging())
		
		r.With(middlerware.ValidateRequest("signup")).Post("/signup", h.User.SignupHandler)
		r.With(middlerware.ValidateRequest("login")).Post("/login", h.User.LoginHandler)
		r.Post("/refresh", h.Auth.RefreshTokenHandler)
		r.Post("/logout", h.Auth.LogoutHandler)
		r.Post("/check-email", h.Auth.CheckEmailHandler)

		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.Get("/me", h.Auth.GetMeHandler)
			r.Get("/validate", h.Auth.ValidateTokenHandler)
			r.With(middlerware.ValidateRequest("default")).Post("/change-password", h.Auth.ChangePasswordHandler)
		})
	})

	// User routes
	r.Route("/api/users", func(r chi.Router) {
		r.Use(middlerware.RateLimit("default"))
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.With(middlerware.ValidateRequest("profile")).Put("/profile", h.User.UpdateProfileHandler)
			r.With(middlerware.RateLimit("search")).Get("/search", h.User.SearchUsersHandler)
		})
	})

	// Friends routes
	r.Route("/api/friends", func(r chi.Router) {
		r.Use(middlerware.RateLimit("friends"))
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)

			r.Get("/", h.Friends.GetUserFriendsHandler)
			r.With(middlerware.ValidateRequest("friend_request")).Post("/request", h.Friends.SendFriendRequestHandler)
			r.Post("/{friendshipID}/accept", h.Friends.AcceptFriendRequestHandler)
			r.Post("/{friendshipID}/reject", h.Friends.RejectFriendRequestHandler)
			r.Delete("/{friendshipID}", h.Friends.RemoveFriendHandler)
			r.With(middlerware.ValidateRequest("default")).Post("/block", h.Friends.BlockUserHandler)
			r.Get("/requests/received", h.Friends.GetPendingFriendRequestsHandler)
			r.Get("/requests/sent", h.Friends.GetSentFriendRequestsHandler)
		})
	})

	// Message routes
	r.Route("/api/messages", func(r chi.Router) {
		r.Use(middlerware.RateLimit("message"))
		
		r.Group(func(r chi.Router) {
			r.Use(middlerware.RequireAuth)
			r.With(middlerware.ValidateRequest("message")).Post("/", h.Message.SendMessageHandler)
			r.Get("/conversations", h.Message.GetConversationsHandler)
			r.With(middlerware.RateLimit("search")).Get("/conversations/search", h.Message.SearchConversationsHandler)
			r.Get("/{userID}", h.Message.GetMessagesHandler)
			r.Put("/read/{userID}", h.Message.MarkAsReadHandler)
			r.Get("/unread/{userID}", h.Message.GetUnreadCountHandler)
			r.Delete("/{messageID}", h.Message.DeleteMessageHandler)
			r.With(middlerware.RateLimit("search")).Get("/search", h.Message.SearchMessagesHandler)
		})
	})

	r.With(middlerware.RateLimit("default")).Get("/ws", h.WebSocket.ServeWS)

	return nil
}
