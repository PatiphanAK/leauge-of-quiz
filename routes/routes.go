package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

// SetupRoutes ตั้งค่าเส้นทางทั้งหมด
func SetupRoutes(app *fiber.App, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Middleware
	app.Use(logger.New())  // สำหรับ log การร้องขอ
	app.Use(recover.New()) // สำหรับ recover จากการ panics

	// routes
	SetupAuthRoute(app, authHandler, authMiddleware)
	// SetupQuizRoute(app, &handlers.QuizHandler{DB: database.DB}, authMiddleware)
}
