package routes

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
	models "github.com/patiphanak/league-of-quiz/model"
)

func SetupAuthRoute(app *fiber.App, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	if authHandler == nil {
		log.Fatal("❌ authHandler is nil!")
	}
	apiV1 := app.Group("/api/v1")

	auth := app.Group("/auth")
	auth.Get("/google", authHandler.GoogleLogin)
	auth.Get("/google/callback", authHandler.GoogleCallback)
	auth.Post("/logout", authHandler.Logout)

	authProtected := apiV1.Group("/auth", authMiddleware.RequireAuth())
	authProtected.Get("/me", func(c *fiber.Ctx) error {
		userLocal := c.Locals("user")
		user, ok := userLocal.(models.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - invalid user format",
			})
		}
		fmt.Printf("Sending user data: %+v\n", user)
	
    return c.JSON(fiber.Map{
        "user": fiber.Map{
            "id":      user.ID,
            "email":   user.Email,
            "name":    user.DisplayName,
            "picture": user.PictureURL,
        },
    })
	})
}