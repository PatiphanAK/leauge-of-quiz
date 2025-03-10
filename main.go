package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/patiphanak/league-of-quiz/auth/jwt"
	"github.com/patiphanak/league-of-quiz/auth/oauth"
	"github.com/patiphanak/league-of-quiz/config"
	"github.com/patiphanak/league-of-quiz/database"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
	"github.com/patiphanak/league-of-quiz/repositories"
	routes "github.com/patiphanak/league-of-quiz/routes"
	"github.com/patiphanak/league-of-quiz/services"
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

	// Set up storage paths
	storageBasePath := "./storage"
	if os.Getenv("STORAGE_PATH") != "" {
		storageBasePath = os.Getenv("STORAGE_PATH")
	}

	// Initialize Repository
	repos := repositories.InitRepositories(database.DB)

	// Initialize Services with proper storage path
	services, err := services.InitServices(repos, storageBasePath)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.ShutdownServices()

	// Initialize auth components
	googleAuth := oauth.NewGoogleOAuth(cfg)
	jwtService := jwt.NewJWTService(cfg)

	// Set up handlers
	authHandler := handlers.NewAuthHandler(database.DB, googleAuth, jwtService)
	quizHandler := handlers.NewQuizHandler(services.Quiz, services.File)
	uploadHandler := handlers.NewUploadHandler(services.File)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Set up middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4000",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(database.DB, jwtService)

	// Set up routes
	routes.SetupRoutes(app, authHandler, authMiddleware, quizHandler, uploadHandler)
	// You may want to add your quiz routes here
	// Example: routes.SetupQuizRoutes(app, quizHandler, authMiddleware)

	// Set up static file server for uploaded files
	// This should point to the same base directory used by FileService
	app.Static("/storage", storageBasePath)

	// Set up graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Start the server
	log.Println("Server starting on port 3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	log.Println("Server stopped")
}
