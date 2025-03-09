package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupQuizRoute(app *fiber.App, quizHandler *handlers.QuizHandler, authMiddleware *middleware.AuthMiddleware) {
	apiV1 := app.Group("/api/v1")
	quizRoutes := apiV1.Group("/quizzes")
	quizRoutes.Get("/", quizHandler.GetQuizzes)
	quizRoutes.Get("/:id", quizHandler.GetQuizByID)
	quizRoutes.Get("/my", authMiddleware.RequireAuth(), quizHandler.GetMyQuizzes)
	quizRoutes.Get("/categories", quizHandler.GetCategories)

	// ต้องมีการตรวจสอบ authentication
	quizRoutes.Use(authMiddleware.RequireAuth())
	quizRoutes.Post("/", quizHandler.CreateQuiz)
	quizRoutes.Post("/form", quizHandler.CreateQuizWithForm)
	quizRoutes.Patch("/:id", quizHandler.PatchQuiz)
	quizRoutes.Delete("/:id", quizHandler.DeleteQuiz)
}
