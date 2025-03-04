package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
	models "github.com/patiphanak/league-of-quiz/model"
)

// SetupRoutes ตั้งค่าเส้นทางทั้งหมด
func SetupRoutes(app *fiber.App, authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	// Middleware
	app.Use(logger.New())  // สำหรับ log การร้องขอ
	app.Use(recover.New()) // สำหรับ recover จากการ panics

	// routes
	SetupAuthRoute(app, authHandler, authMiddleware)
	SetupQuizRoute(app, &handlers.QuizHandler{}, authMiddleware)

	// Protected routes (ที่ต้องใช้การยืนยันตัวตน)
	api := app.Group("/api", authMiddleware.RequireAuth()) // ใช้ middleware สำหรับการตรวจสอบการยืนยันตัวตน

	// เส้นทางสำหรับข้อมูลของผู้ใช้
	api.Get("/me", func(c *fiber.Ctx) error {
		// ตรวจสอบว่า user มีค่าหรือไม่ใน context
		userLocal := c.Locals("user")
		if userLocal == nil {
			// ส่งข้อความตอบกลับหากไม่พบ user ใน context
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - no user found in context",
			})
		}

		// แปลง userLocal เป็น *models.User
		user, ok := userLocal.(*models.User)
		if !ok {
			// ส่งข้อความตอบกลับหากไม่สามารถแปลง user เป็น *models.User
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - invalid user format",
			})
		}

		// ส่งข้อมูลของผู้ใช้กลับ
		return c.JSON(fiber.Map{
			"id":      user.ID,
			"email":   user.Email,
			"name":    user.DisplayName,
			"picture": user.PictureURL,
		})
	})
}
