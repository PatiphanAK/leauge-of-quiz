package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middlewares "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupQuizRoute(app *fiber.App, quizHandler *handlers.QuizHandler, authMiddleware *middlewares.AuthMiddleware) {
	if quizHandler == nil {
		log.Fatal("❌ quizHandler is nil!")
	}

	// กลุ่ม route สำหรับ Quiz ที่ต้องมีการยืนยันตัวตน
	quiz := app.Group("/api/quizzes", authMiddleware.RequireAuth())

	// เส้นทางสำหรับ Quiz
	quiz.Get("/", quizHandler.GetQuizzes)
	quiz.Get("/:id", quizHandler.GetQuizByID)
	quiz.Post("/", quizHandler.CreateQuiz)
	quiz.Put("/:id", quizHandler.UpdateQuiz)
	quiz.Patch("/:id", quizHandler.UpdateQuiz)
	quiz.Delete("/:id", quizHandler.DeleteQuiz)
}
