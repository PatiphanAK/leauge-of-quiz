package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	jwt "github.com/patiphanak/league-of-quiz/auth/jwt"
	"github.com/patiphanak/league-of-quiz/auth/oauth"
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db          *gorm.DB
	googleOAuth *oauth.GoogleOAuth
	jwtService  *jwt.JWTService
}

func NewAuthHandler(db *gorm.DB, googleOAuth *oauth.GoogleOAuth, jwtService *jwt.JWTService) *AuthHandler {
	return &AuthHandler{
		db:          db,
		googleOAuth: googleOAuth,
		jwtService:  jwtService,
	}
}

// Google login ส่ง redirect ไป Google OAuth Login
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state := uuid.New().String()

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
	})

	authURL := h.googleOAuth.GetAuthURL(state)
	return c.Redirect(authURL)
}

// GoogleCallback
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	state := c.Cookies("oauth_state")
	code := c.Query("code")

	// ตรวจสอบ state ว่าตรงกันหรือไม่
	cookie := c.Cookies("oauth_state")
	if cookie == "" || cookie != state {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state parameter",
		})
	}

	// แลกเปลี่ยน authorization code เป็น access token
	token, err := h.googleOAuth.Exchange(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	// ดึงข้อมูลผู้ใช้จาก Google
	userInfo, err := h.googleOAuth.GetUserInfo(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}

	// ค้นหาผู้ใช้ในฐานข้อมูล
	var user models.User
	result := h.db.Where("google_id = ?", userInfo.ID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// ผู้ใช้ยังไม่มีในฐานข้อมูล
			user = models.User{
				GoogleID:    userInfo.ID,
				Email:       userInfo.Email,
				DisplayName: userInfo.Name,
				PictureURL:  userInfo.Picture,
			}

			// สร้างผู้ใช้ใหม่
			if err := h.db.Create(&user).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create user",
				})
			}
		} else {
			// ถ้ามีข้อผิดพลาดในการค้นหาผู้ใช้
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to query user from database",
			})
		}
	}

	// สร้าง JWT token
	jwtToken, err := h.jwtService.GenerateToken(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// ตั้งค่า cookie แบบ HTTP-only เพื่อป้องกัน XSS
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		HTTPOnly: true,
		// ถ้าใช้งานบน localhost ในระหว่างการพัฒนา ให้ตั้งค่า Secure เป็น false
		// ในสภาพแวดล้อมการทำงานจริง (production) ให้เปลี่ยนเป็น true
		Secure: false,
		// SameSite: "Lax" มีความเหมาะสมกับ redirects มากกว่า "Strict"
		SameSite: "Lax",
		MaxAge:   60 * 60 * 24 * 7, // 1 week
	})

	// ส่งข้อมูลกลับโดยไม่รวม token ใน response body
	// เนื่องจากเราใช้ HTTP-only cookie แล้ว
	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"pictureUrl":  user.PictureURL,
		},
	})
}

// Logout จัดการการออกจากระบบโดยลบ cookie
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// ลบ cookie โดยการตั้งค่า MaxAge เป็นค่าลบ
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false, // ตั้งเป็น true ในโหมด production
	})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}
