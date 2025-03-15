package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupRoutes(app *fiber.App, handlers *handlers.AllHandlers, authMiddleware *middleware.AuthMiddleware) {
	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.TransformResponse())

	// routes
	SetupAuthRoute(app, handlers.Auth, authMiddleware)
	SetupUploadRoutes(app, handlers.Upload, authMiddleware)
	SetupQuizRoute(app, handlers.Quiz, authMiddleware)
	SetupQuestionRoute(app, handlers.Question, authMiddleware)
}

func SetupGameRoutes(app *fiber.App, gameHandler *handlers.GameHandler, authMiddleware *middleware.AuthMiddleware) {
	// Regular HTTP routes
	apiV1 := app.Group("/api/v1")
	gameRoutes := apiV1.Group("/games")

	// Public routes
	gameRoutes.Get("/sessions/:id", gameHandler.GetGameSession)

	// Protected routes
	gameRoutes.Use(authMiddleware.RequireAuth())
	gameRoutes.Post("/sessions", gameHandler.CreateGameSession)
	gameRoutes.Post("/sessions/join", gameHandler.JoinGameSession)

	// Hook up Socket.IO
	app.Use(socketio.New(socketio.Config{
		Server: gameHandler.GetSocketIOServer(),
	}))

	// WebSocket endpoint
	app.Get("/socket.io/*", func(c *fiber.Ctx) error {
		// This route is handled by the Socket.IO middleware
		return nil
	})
}
