package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
)

func SetupAuthRoute(app *fiber.App, authHandler *handlers.AuthHandler) {
	if authHandler == nil {
		log.Fatal("❌ authHandler is nil!")
	}

	auth := app.Group("/auth")
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
}
