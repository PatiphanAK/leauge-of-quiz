package handlers

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

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

	// เรียกใช้ service
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

	quiz, err := h.quizService.GetQuizByID(quizID)
	if err != nil {
		// เช็คว่า error มาจากการไม่พบ Quiz หรือ Database Error
		if errors.Is(err, gorm.ErrRecordNotFound) { // ใช้กับ GORM
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Quiz not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{"data": quiz})
}

// CreateQuiz สร้าง quiz ใหม่ (รองรับ multipart form)
func (h *QuizHandler) CreateQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับข้อมูลจาก multipart form
	title := c.FormValue("title")
	description := c.FormValue("description")
	timeLimit, _ := strconv.Atoi(c.FormValue("timeLimit", "0"))
	isPublished := c.FormValue("isPublished") == "true"
	
	// ตรวจสอบข้อมูลที่จำเป็น
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}

	// รับไฟล์รูปภาพ (ถ้ามี)
	imageFile, err := c.FormFile("image")
	
	// สร้าง quiz object
	quiz := &models.Quiz{
		Title:       title,
		Description: description,
		TimeLimit:   uint(timeLimit),
		IsPublished: isPublished,
		CreatorID:   userID,
	}

	// สร้าง quiz พร้อมรูปภาพ
	if err := h.quizService.CreateQuiz(quiz, imageFile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// รับข้อมูลหมวดหมู่ (ถ้ามี)
	categoriesStr := c.FormValue("categories")
	if categoriesStr != "" {
		var categoryIDs []uint
		categoryStrings := strings.Split(categoriesStr, ",")
		for _, catStr := range categoryStrings {
			catID, err := strconv.ParseUint(catStr, 10, 32)
			if err == nil {
				categoryIDs = append(categoryIDs, uint(catID))
			}
		}
		
		// อัปเดตหมวดหมู่
		if len(categoryIDs) > 0 {
			if err := h.quizService.UpdateQuizCategories(quiz.ID, categoryIDs, userID); err != nil {
				// ถ้ามีข้อผิดพลาดในการอัปเดตหมวดหมู่ ก็ไม่ต้องยกเลิกการสร้าง quiz
				log.Printf("Failed to update quiz categories: %v", err)
			}
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Quiz created successfully",
		"data": fiber.Map{
			"id": quiz.ID,
		},
	})
}

// UpdateQuiz อัปเดต quiz (รองรับ multipart form)
func (h *QuizHandler) UpdateQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
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
	
	// รับข้อมูลจาก multipart form
	title := c.FormValue("title", quiz.Title)
	log.Printf("title: %v", title)
	description := c.FormValue("description", quiz.Description)
	timeLimitStr := c.FormValue("timeLimit")
	isPublishedStr := c.FormValue("isPublished")
	
	// ตรวจสอบข้อมูลที่จำเป็น
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Title is required"})
	}
	if description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Description is required"})
	}
	
	// สร้าง map สำหรับการอัปเดต
	updates := map[string]interface{}{
		"title":       title,
		"description": description,
	}
	
	// เพิ่มข้อมูลเพิ่มเติมถ้ามีการส่งมา
	if timeLimitStr != "" {
		timeLimit, _ := strconv.Atoi(timeLimitStr)
		updates["time_limit"] = timeLimit
	}
	
	if isPublishedStr != "" {
		isPublished := isPublishedStr == "true"
		updates["is_published"] = isPublished
	}
	
	// รับไฟล์รูปภาพ (ถ้ามี)
	imageFile, _ := c.FormFile("imageURL")
	
	// อัปเดตข้อมูล quiz
	if err := h.quizService.PatchQuiz(quizID, updates, imageFile, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	
	// อัปเดตหมวดหมู่ (ถ้ามี)
	categoriesStr := c.FormValue("categories")
	if categoriesStr != "" {
		var categoryIDs []uint
		categoryStrings := strings.Split(categoriesStr, ",")
		for _, catStr := range categoryStrings {
			catID, err := strconv.ParseUint(catStr, 10, 32)
			if err == nil {
				categoryIDs = append(categoryIDs, uint(catID))
			}
		}
		
		if err := h.quizService.UpdateQuizCategories(quizID, categoryIDs, userID); err != nil {
			// ถ้ามีข้อผิดพลาดในการอัปเดตหมวดหมู่ ก็ไม่ต้องยกเลิกการอัปเดต quiz
			log.Printf("Failed to update quiz categories: %v", err)
		}
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
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ ID จาก parameter
	quizID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ลบ quiz
	if err := h.quizService.DeleteQuiz(quizID, userID); err != nil {
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
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
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

// GetCategories ดึงข้อมูลหมวดหมู่ทั้งหมด
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