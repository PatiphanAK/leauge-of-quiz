package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/auth/jwt"
	"github.com/patiphanak/league-of-quiz/auth/oauth"
	"github.com/patiphanak/league-of-quiz/config"
	"github.com/patiphanak/league-of-quiz/database"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
	"github.com/patiphanak/league-of-quiz/routes"
)

func main() {
	// Load environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	database.InitDB()
	database.AutoMigration(database.DB)

	// Initialize OAuth provider
	googleAuth := oauth.NewGoogleOAuth(cfg)

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg)

	// Create an instance of AuthHandler with dependencies
	authHandler := handlers.NewAuthHandler(database.DB, googleAuth, jwtService)

	// Create Fiber app
	app := fiber.New()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(database.DB, jwtService)

	// Setup routes
	routes.SetupRoutes(app, authHandler, authMiddleware)

	// Start the server
	log.Println("Server starting on port 3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
