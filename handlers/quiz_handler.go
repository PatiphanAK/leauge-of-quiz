package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/services"
)

type QuizHandler struct {
	quizService *services.QuizService
}

// NewQuizHandler สร้าง instance ใหม่ของ QuizHandler
func NewQuizHandler(quizService *services.QuizService) *QuizHandler {
	log.Println("NewQuizHandler")
	return &QuizHandler{quizService: quizService}
}

// GetQuizzes ดึงข้อมูล quizzes ทั้งหมด
func (h *QuizHandler) GetQuizzes(c *fiber.Ctx) error {
	// รับ pagination parameters
	log.Println("get api/v1/quizzes")
	page, _ := strconv.Atoi(c.Query("page", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// ตรวจสอบว่าต้องการเฉพาะ published quizzes หรือไม่
	publishedOnly := c.Query("published", "false") == "true"

	var quizzes []models.Quiz
	var count int64
	var err error

	if publishedOnly {
		quizzes, count, err = h.quizService.GetPublishedQuizzes(page, limit)
	} else {
		quizzes, count, err = h.quizService.GetAllQuizzes(page, limit)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": quizzes,
		"meta": fiber.Map{
			"total": count,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetQuizByID ดึงข้อมูล quiz จาก ID
func (h *QuizHandler) GetQuizByID(c *fiber.Ctx) error {
	// รับ ID จาก parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	quiz, err := h.quizService.GetQuizByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Quiz not found",
		})
	}

	return c.JSON(fiber.Map{
		"data": quiz,
	})
}

// CreateQuiz สร้าง quiz ใหม่
func (h *QuizHandler) CreateQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to create a quiz",
		})
	}

	// รับข้อมูลจาก request body
	var request struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		TimeLimit   uint   `json:"timeLimit"`
		IsPublished bool   `json:"isPublished"`
		ImageURL    string `json:"imageURL"`
		Categories  []uint `json:"categories"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if request.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if request.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}

	// สร้าง quiz object
	quiz := &models.Quiz{
		Title:       request.Title,
		Description: request.Description,
		TimeLimit:   request.TimeLimit,
		IsPublished: request.IsPublished,
		ImageURL:    request.ImageURL,
		CreatorID:   userID,
	}

	// บันทึกข้อมูล
	if err := h.quizService.CreateQuiz(quiz); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// อัปเดตหมวดหมู่ (ถ้ามี)
	if len(request.Categories) > 0 {
		if err := h.quizService.UpdateQuizCategories(quiz.ID, request.Categories, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Quiz created successfully",
		"data": fiber.Map{
			"id": quiz.ID,
		},
	})
}

// PatchQuiz อัปเดตข้อมูล quiz บางส่วน
func (h *QuizHandler) PatchQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to update a quiz",
		})
	}

	// รับ ID จาก parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// รับข้อมูลจาก request body
	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// ตรวจสอบว่ามีการส่ง categories มาหรือไม่
	var categories []uint
	if categoriesInterface, exists := updates["categories"]; exists {
		delete(updates, "categories")

		// แปลง interface{} เป็น []uint
		if categoriesArray, ok := categoriesInterface.([]interface{}); ok {
			for _, category := range categoriesArray {
				if categoryFloat, ok := category.(float64); ok {
					categories = append(categories, uint(categoryFloat))
				}
			}
		}
	}

	// อัปเดตข้อมูล quiz
	if err := h.quizService.PatchQuiz(uint(id), updates, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// อัปเดตหมวดหมู่ (ถ้ามี)
	if len(categories) > 0 {
		if err := h.quizService.UpdateQuizCategories(uint(id), categories, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Quiz updated successfully",
	})
}

// DeleteQuiz ลบ quiz
func (h *QuizHandler) DeleteQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to delete a quiz",
		})
	}

	// รับ ID จาก parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// ลบ quiz
	if err := h.quizService.DeleteQuiz(uint(id), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Quiz deleted successfully",
	})
}

// GetMyQuizzes ดึงข้อมูล quizzes ที่สร้างโดยผู้ใช้ปัจจุบัน
func (h *QuizHandler) GetMyQuizzes(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to view your quizzes",
		})
	}

	// รับ pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	quizzes, count, err := h.quizService.GetQuizzesByCreator(userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": quizzes,
		"meta": fiber.Map{
			"total": count,
			"page":  page,
			"limit": limit,
		},
	})
}
