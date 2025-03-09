package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupQuizRoute(app *fiber.App, quizHandler *handlers.QuizHandler, authMiddleware *middleware.AuthMiddleware) {
	apiV1 := app.Group("/api/v1")
	quizRoutes := apiV1.Group("/quiz")
	quizRoutes.Get("/", quizHandler.GetQuizzes)
	quizRoutes.Get("/:id", quizHandler.GetQuizByID)
	quizRoutes.Post("/", quizHandler.CreateQuiz)
	quizRoutes.Patch("/:id", quizHandler.PatchQuiz)
	quizRoutes.Delete("/:id", quizHandler.DeleteQuiz)
}