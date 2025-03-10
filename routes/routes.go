package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

// SetupRoutes ตั้งค่าเส้นทางทั้งหมด
func SetupRoutes(app *fiber.App, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware, quizHandler *handlers.QuizHandler, uploadHandler *handlers.UploadHandler) {
	// Middleware
	app.Use(logger.New())  // สำหรับ log การร้องขอ
	app.Use(recover.New()) // สำหรับ recover จากการ panics
	app.Use(middleware.TransformResponse())

	// routes
	SetupAuthRoute(app, authHandler, authMiddleware)
	SetupUploadRoutes(app, uploadHandler, authMiddleware)
	SetupQuizRoute(app, quizHandler, authMiddleware)
}
