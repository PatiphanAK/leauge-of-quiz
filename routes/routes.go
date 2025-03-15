package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
	"github.com/patiphanak/league-of-quiz/websocket"
)

func SetupRoutes(app *fiber.App, handlers *handlers.AllHandlers, authMiddleware *middleware.AuthMiddleware, wsManager *websocket.Manager) {
	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(middleware.TransformResponse())

	// routes
	SetupAuthRoute(app, handlers.Auth, authMiddleware)
	SetupUploadRoutes(app, handlers.Upload, authMiddleware)
	SetupQuizRoute(app, handlers.Quiz, authMiddleware)
	SetupQuestionRoute(app, handlers.Question, authMiddleware)
	SetupGameRoute(app, handlers.Game, authMiddleware)
	SetupWebSocketRoute(app, wsManager)
}
