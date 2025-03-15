package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

// SetupGameRoute ลงทะเบียน routes สำหรับเกม
func SetupGameRoute(app *fiber.App, gameHandler *handlers.GameHandler, authMiddleware *middleware.AuthMiddleware) {
	// กลุ่ม API endpoints สำหรับเกม
	apiV1 := app.Group("/api/v1")

	// เส้นทางที่ต้องการการยืนยันตัวตน
	gameAPI := apiV1.Group("/games", authMiddleware.RequireAuth())

	// จัดการ session
	gameAPI.Post("/sessions", gameHandler.CreateGameSession)
	gameAPI.Get("/sessions", gameHandler.GetGameSessions)
	gameAPI.Get("/sessions/:id", gameHandler.GetGameSessionDetail)
	gameAPI.Post("/sessions/:id/join", gameHandler.JoinGameSession)
	gameAPI.Post("/sessions/:id/start", gameHandler.StartGameSession)
	gameAPI.Post("/sessions/:id/end", gameHandler.EndGameSession)

	// จัดการคำตอบ
	gameAPI.Post("/sessions/:id/answers", gameHandler.SubmitAnswer)

	// ดูผลลัพธ์
	gameAPI.Get("/sessions/:id/results", gameHandler.GetGameResults)
}
