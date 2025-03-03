package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

func SetupAuthRoute(app *fiber.App, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	if authHandler == nil {
		log.Fatal("❌ authHandler is nil!")
	}

	auth := app.Group("/auth")
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	auth.Post("/logout", authHandler.Logout)

	// เปลี่ยนจาก authHandler.Me เป็น authHandler.GetCurrentUser
	// และเพิ่ม parameter authMiddleware
	app.Get("/api/me", authMiddleware.RequireAuth(), authHandler.GetCurrentUser)
}