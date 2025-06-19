package main

import (
	"fmt"
	"log"
	"net/http"

	"go-chat/config"
	"go-chat/internal/domain"
	"go-chat/internal/handlers"
	"go-chat/internal/routes"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	db, err := config.ConnectToDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run auto-migration
	if err := db.AutoMigrate(&domain.User{}, &domain.Friendship{}, &domain.Message{}); err != nil {
		log.Fatal("Failed to auto-migrate database:", err)
	}

	fmt.Println("Database migration completed successfully")

	// Initialize handlers
	h := handlers.NewHandlers(db)

	// Setup router
	r := chi.NewRouter()

	// Setup routes
	if err := routes.SetupRoutes(r, db, h); err != nil {
		log.Fatal("Failed to setup routes:", err)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server starting on port %s\n", cfg.Port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
