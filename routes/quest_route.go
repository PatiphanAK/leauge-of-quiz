package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupQuestionRoute(app *fiber.App, questionHandler *handlers.QuestionHandler, authMiddleware *middleware.AuthMiddleware) {
	apiV1 := app.Group("/api/v1")
	quizzes := apiV1.Group("/quizzes")
	
	// Routes that don't require auth
	quizzes.Get("/:quizId/questions", questionHandler.GetQuestionsByQuizID)
	quizzes.Get("/:quizId/questions/:id", questionHandler.GetQuestionByID)
	
	// Routes that require auth
	questionsWithAuth := quizzes.Group("/:quizId/questions")
	questionsWithAuth.Use(authMiddleware.RequireAuth())
	questionsWithAuth.Post("/", questionHandler.CreateQuestion)
	questionsWithAuth.Put("/:id", questionHandler.UpdateQuestion)
	questionsWithAuth.Patch("/:id", questionHandler.PatchQuestion)
	questionsWithAuth.Delete("/:id", questionHandler.DeleteQuestion)
}