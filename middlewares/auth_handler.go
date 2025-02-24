// middleware/auth_middleware.go
package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/auth/jwt"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	db         *gorm.DB
	jwtService *jwt.JWTService
}

func NewAuthMiddleware(db *gorm.DB, jwtService *jwt.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		db:         db,
		jwtService: jwtService,
	}
}

// RequireAuth middleware ที่ตรวจสอบว่าผู้ใช้ได้เข้าสู่ระบบหรือไม่
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึง Authorization header
		auth := c.Get("Authorization")

		// ตรวจสอบว่ามี Bearer token หรือไม่
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// แยก token ออกจาก Bearer prefix
		tokenString := strings.TrimPrefix(auth, "Bearer ")

		// ตรวจสอบความถูกต้องของ token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// ดึงข้อมูลผู้ใช้จากฐานข้อมูล
		var user models.User
		if err := m.db.First(&user, claims.UserID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// เก็บข้อมูลผู้ใช้ใน locals ของ context
		c.Locals("user", user)

		// ดำเนินการต่อไปยัง handler ถัดไป
		return c.Next()
	}
}
