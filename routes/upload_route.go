package routes

import (
	"github.com/gofiber/fiber/v2"
	handler "github.com/patiphanak/league-of-quiz/handlers"
	middleware "github.com/patiphanak/league-of-quiz/middlewares"
)

// SetupUploadRoutes กำหนด routes สำหรับการอัปโหลดไฟล์
func SetupUploadRoutes(app *fiber.App, uploadHandler *handler.UploadHandler, authMiddleware *middleware.AuthMiddleware) {
	// กำหนด route group สำหรับการอัปโหลด
	uploadRoutes := app.Group("/api/upload")

	// ต้องล็อกอินก่อนถึงจะอัปโหลดได้
	uploadRoutes.Use(authMiddleware.RequireAuth())

	// Route เดียวสำหรับการอัปโหลดไฟล์ทุกประเภท โดยใช้ path parameter
	uploadRoutes.Post("/:type", uploadHandler.UploadFile)

	// Route เดียวสำหรับการลบไฟล์ทุกประเภท
	uploadRoutes.Delete("/:type/:filename", uploadHandler.DeleteFile)
}
