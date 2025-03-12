package handlers

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/patiphanak/league-of-quiz/dto"
	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/services"
	"github.com/patiphanak/league-of-quiz/utils"
)

type QuizHandler struct {
	quizService *services.QuizService
	fileService *services.FileService
}

func NewQuizHandler(quizService *services.QuizService, fileService *services.FileService) *QuizHandler {
	log.Println("NewQuizHandler")
	return &QuizHandler{
		quizService: quizService,
		fileService: fileService,
	}
}

// GetQuizzes ดึงข้อมูล quizzes ทั้งหมด
func (h *QuizHandler) GetQuizzes(c *fiber.Ctx) error {
	// รับ pagination parameters
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	// รับ filter parameters
	isPublished := c.Query("isPublished", "")
	search := c.Query("search", "")

	// รับ categories (อาจเป็น comma-separated values)
	categoriesStr := c.Query("categories", "")
	var categories []uint
	if categoriesStr != "" {
		categoryStrings := strings.Split(categoriesStr, ",")
		for _, catStr := range categoryStrings {
			catID, err := strconv.ParseUint(catStr, 10, 32)
			if err == nil {
				categories = append(categories, uint(catID))
			}
		}
	}

	var quizzes []models.Quiz
	var count int64
	var err error

	// เรียกใช้ service แบบใหม่
	quizzes, count, err = h.quizService.GetFilteredQuizzes(offset, limit, isPublished, search, categories)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": quizzes,
		"meta": fiber.Map{
			"total":  count,
			"offset": offset,
			"limit":  limit,
		},
	})
}

// GetQuizByID ดึงข้อมูล quiz จาก ID
func (h *QuizHandler) GetQuizByID(c *fiber.Ctx) error {
	quizID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		if statusCode == fiber.StatusOK { // ป้องกัน statusCode 200 ผิดที่
			statusCode = fiber.StatusBadRequest
		}
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	quiz, err := h.quizService.GetQuizByID(uint(quizID))
	if err != nil {
		// เช็คว่า error มาจากการไม่พบ Quiz หรือ Database Error
		if errors.Is(err, gorm.ErrRecordNotFound) { // ใช้กับ GORM
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Quiz not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{"data": quiz})
}

func (h *QuizHandler) CreateQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับข้อมูลจาก request body
	var request dto.CreateQuizRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	log.Printf("request: %v", request)
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

	// สร้าง quiz พร้อมคำถามและตัวเลือก
	if err := h.quizService.CreateQuiz(quiz); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Quiz created successfully",
		"data": fiber.Map{
			"id": quiz.ID,
		},
	})
}

// UpdateQuiz อัปเดต quiz
func (h *QuizHandler) UpdateQuiz(c *fiber.Ctx) error {
    // ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	log.Println("Content-Type:", c.Get("Content-Type"))
	log.Println("Raw Body:", string(c.Body()))
	title := c.FormValue("Title")
	title2 := c.FormValue("title")
	log.Println(title)
	log.Println(title2)
    userID, statusCode, err := utils.GetAuthenticatedUserID(c)
    if err != nil {
        return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
    }
    
    quizID, statusCode, err := utils.ParseIDParam(c, "id")
    if err != nil {
        if statusCode == fiber.StatusOK {
            statusCode = fiber.StatusBadRequest
        }
        return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
    }

    // ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
    quiz, err := h.quizService.GetQuizByID(quizID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Quiz not found"})
    }
    
    if quiz.CreatorID != userID {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You don't have permission to modify this quiz"})
    }
    
    // รับข้อมูลจาก request body
    var request dto.UpdateQuizRequest
    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }
    
    // ตรวจสอบข้อมูลที่จำเป็น
    if request.Title == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
    }
    if request.Description == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Description is required"})
    }
    
    // 1. อัปเดตข้อมูลพื้นฐานของ quiz
    updates := map[string]interface{}{
        "title": request.Title,
        "description": request.Description,
        "time_limit": request.TimeLimit,
        "is_published": request.IsPublished,
    }
    
    if err := h.quizService.PatchQuiz(quizID, updates, userID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    
    // 2. อัปเดตหมวดหมู่
    if err := h.quizService.UpdateQuizCategories(quizID, request.Categories, userID); err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    
    // ดึงข้อมูล quiz ที่อัปเดตแล้วและส่งกลับ
    updatedQuiz, err := h.quizService.GetQuizByID(quizID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to retrieve updated quiz",
        })
    }
    
    return c.JSON(fiber.Map{
        "message": "Quiz updated successfully",
        "data": updatedQuiz,
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

func (h *QuizHandler) GetCategories(c *fiber.Ctx) error {
	// ดึงข้อมูลหมวดหมู่ทั้งหมดจากฐานข้อมูล
	categories, err := h.quizService.GetAllCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

