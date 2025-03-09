package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/services"
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
		Questions   []struct {
			Text     string `json:"text"`
			ImageURL string `json:"imageURL"`
			Choices  []struct {
				Text      string `json:"text"`
				ImageURL  string `json:"imageURL"`
				IsCorrect bool   `json:"isCorrect"`
			} `json:"choices"`
		} `json:"questions"`
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

	// แปลง request.Questions เป็น []services.QuestionData
	questions := make([]services.QuestionData, len(request.Questions))
	for i, q := range request.Questions {
		questions[i].Text = q.Text
		questions[i].ImageURL = q.ImageURL

		questions[i].Choices = make([]services.ChoiceData, len(q.Choices))
		for j, c := range q.Choices {
			questions[i].Choices[j].Text = c.Text
			questions[i].Choices[j].ImageURL = c.ImageURL
			questions[i].Choices[j].IsCorrect = c.IsCorrect
		}
	}

	// สร้าง quiz พร้อมคำถามและตัวเลือก
	if err := h.quizService.CreateQuizWithQuestionsAndChoices(quiz, questions, request.Categories, userID); err != nil {
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

// CreateQuizWithForm สร้าง quiz ใหม่พร้อมคำถามและตัวเลือกจาก FormData
func (h *QuizHandler) CreateQuizWithForm(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to create a quiz",
		})
	}

	// รับข้อมูล quiz จาก FormData
	quizDataStr := c.FormValue("quizData", "{}")
	var quizData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		TimeLimit   uint   `json:"timeLimit"`
		IsPublished bool   `json:"isPublished"`
		Categories  []uint `json:"categories"`
		Questions   []struct {
			Text    string `json:"text"`
			Choices []struct {
				Text      string `json:"text"`
				IsCorrect bool   `json:"isCorrect"`
			} `json:"choices"`
		} `json:"questions"`
	}

	if err := json.Unmarshal([]byte(quizDataStr), &quizData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz data format",
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if quizData.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	if quizData.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}

	// อัปโหลดรูปภาพของ quiz (ถ้ามี)
	var quizImageURL string
	quizImage, err := c.FormFile("quizImage")
	if err == nil && quizImage != nil {
		quizImageURL, err = h.fileService.UploadFile(quizImage, "quiz")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload quiz image: " + err.Error(),
			})
		}
	}

	// สร้าง quiz
	quiz := &models.Quiz{
		Title:       quizData.Title,
		Description: quizData.Description,
		TimeLimit:   quizData.TimeLimit,
		IsPublished: quizData.IsPublished,
		ImageURL:    quizImageURL,
		CreatorID:   userID,
	}

	// เตรียมข้อมูลคำถามที่มีรูปภาพ
	questionsWithImages := make([]services.QuestionData, len(quizData.Questions))

	// อัปโหลดรูปภาพของคำถามและตัวเลือก
	for i, q := range quizData.Questions {
		// เก็บข้อมูลคำถาม
		questionsWithImages[i].Text = q.Text

		// อัปโหลดรูปภาพคำถาม (ถ้ามี)
		questionImageField := fmt.Sprintf("questionImage_%d", i)
		questionImage, err := c.FormFile(questionImageField)
		if err == nil && questionImage != nil {
			questionImageURL, err := h.fileService.UploadFile(questionImage, "question")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to upload question image %d: %s", i, err.Error()),
				})
			}
			questionsWithImages[i].ImageURL = questionImageURL
		}

		// เตรียมตัวเลือก
		questionsWithImages[i].Choices = make([]services.ChoiceData, len(q.Choices))
		for j, choice := range q.Choices {
			questionsWithImages[i].Choices[j].Text = choice.Text
			questionsWithImages[i].Choices[j].IsCorrect = choice.IsCorrect

			// อัปโหลดรูปภาพตัวเลือก (ถ้ามี)
			choiceImageField := fmt.Sprintf("choiceImage_%d_%d", i, j)
			choiceImage, err := c.FormFile(choiceImageField)
			if err == nil && choiceImage != nil {
				choiceImageURL, err := h.fileService.UploadFile(choiceImage, "choice")
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to upload choice image %d-%d: %s", i, j, err.Error()),
					})
				}
				questionsWithImages[i].Choices[j].ImageURL = choiceImageURL
			}
		}
	}

	// ใช้ service สร้าง quiz พร้อมคำถามและตัวเลือก
	if err := h.quizService.CreateQuizWithQuestionsAndChoices(quiz, questionsWithImages, quizData.Categories, userID); err != nil {
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
