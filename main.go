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
	"github.com/patiphanak/league-of-quiz/websocket"
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

	wsManager, err := websocket.NewManager(services.GameService)
	if err != nil {
		log.Fatalf("Error creating WebSocket manager: %v", err)
	}
	// Initialize auth components
	googleAuth := oauth.NewGoogleOAuth(cfg)
	jwtService := jwt.NewJWTService(cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	// Set up middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:4000,http://localhost:3000",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Content-Length,Accept-Language,Accept-Encoding,Connection,Access-Control-Allow-Origin,Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(database.DB, jwtService)

	// Set up routes
	allHandlers := handlers.InitHandlers(services, database.DB, jwtService, googleAuth)
	routes.SetupRoutes(app, allHandlers, authMiddleware, wsManager)

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
