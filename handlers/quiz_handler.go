package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
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

// UpdateQuiz อัปเดต quiz
func (h *QuizHandler) UpdateQuiz(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}
	quizID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		if statusCode == fiber.StatusOK { // ป้องกัน statusCode 200 ผิดที่
			statusCode = fiber.StatusBadRequest
		}
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	quiz, err := h.quizService.GetQuizByID(uint(quizID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Quiz not found",
		})
	}
	if (quiz.CreatorID != userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to modify this quiz",
		})
	}
	request := dto.UpdateQuizRequest{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}
	updates := make(map[string]interface{})
	updates["title"] = request.Title
	updates["description"] = request.Description
	updates["time_limit"] = request.TimeLimit
	updates["is_published"] = request.IsPublished
	// อัปโหลดรูปภาพของ quiz (ถ้ามี)
	quizImage, err := c.FormFile("ImageURL")
	if err == nil && quizImage != nil {
		// ถ้ามีรูปภาพเดิม ให้ลบก่อน
		if quiz.ImageURL != "" {
			filePath, fileType, err := h.fileService.GetFilePath(quiz.ImageURL)
			if err == nil {
				_ = h.fileService.DeleteFile(filePath, fileType)
			}
		}
		// อัปโหลดรูปภาพใหม่
		quizImageURL, err := h.fileService.UploadFile(quizImage, "quiz")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload quiz image: " + err.Error(),
			})
		}
		updates["image_url"] = quizImageURL
	}
	if len(updates) > 0 {
		if err := h.quizService.PatchQuiz(uint(quizID), updates, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// Update categories separately if provided
	if len(request.Categories) > 0 {
		if err := h.quizService.UpdateQuizCategories(uint(quizID), request.Categories, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}
	return c.JSON(quiz)
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
	log.Printf(string(c.Body()))
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to create a quiz",
		})
	}

	// รับข้อมูล quiz จาก FormData
	quizDataStr := c.FormValue("quizData", "{}")
	var quizData dto.QuizFormData

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

// PatchQuizWithForm อัปเดต quiz พร้อมคำถามและตัวเลือกจาก FormData
func (h *QuizHandler) PatchQuizWithForm(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	log.Printf(string(c.Body()))
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "You must be logged in to update a quiz",
		})
	}

	// รับ ID จาก parameter
	quizID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	quiz, err := h.quizService.GetQuizByID(uint(quizID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Quiz not found",
		})
	}

	if quiz.CreatorID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not the owner of this quiz",
		})
	}

	// รับข้อมูล quiz จาก FormData
	quizDataJSON := c.FormValue("quizData")
    if quizDataJSON == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Missing quiz data",
        })
    }

	var quizData struct {
		Title       string `json:"Title"`
		Description string `json:"Description"`
		TimeLimit   uint   `json:"TimeLimit"`
		IsPublished bool   `json:"IsPublished"`
		Categories  []uint `json:"Categories"`
		Questions   []struct {
			ID      uint   `json:"Id"`       // ID สำหรับคำถามที่มีอยู่แล้ว
			Text    string `json:"Text"`
			Choices []struct {
				ID        uint   `json:"Id"`        // ID สำหรับตัวเลือกที่มีอยู่แล้ว
				Text      string `json:"Text"`
				IsCorrect bool   `json:"IsCorrect"`
			} `json:"Choices"`
		} `json:"Questions"`
	}
	if err := json.Unmarshal([]byte(quizDataJSON), &quizData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz data format: " + err.Error(),
		})
	}
	log.Printf("quizData: %v", quizData)

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

	// สร้าง map สำหรับอัปเดต quiz
	updates := make(map[string]interface{})
	updates["title"] = quizData.Title
	updates["description"] = quizData.Description
	updates["time_limit"] = quizData.TimeLimit
	updates["is_published"] = quizData.IsPublished

	// อัปโหลดรูปภาพของ quiz (ถ้ามี)
	quizImage, err := c.FormFile("quizImage")
	if err == nil && quizImage != nil {
		// ถ้ามีรูปภาพเดิม ให้ลบก่อน
		if quiz.ImageURL != "" {
			filePath, fileType, err := h.fileService.GetFilePath(quiz.ImageURL)
			if err == nil {
				_ = h.fileService.DeleteFile(filePath, fileType)
			}
		}

		// อัปโหลดรูปภาพใหม่
		quizImageURL, err := h.fileService.UploadFile(quizImage, "quiz")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload quiz image: " + err.Error(),
			})
		}
		updates["image_url"] = quizImageURL
	}

	// อัปเดต quiz
	if err := h.quizService.PatchQuiz(uint(quizID), updates, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// อัปเดตหมวดหมู่
	if err := h.quizService.UpdateQuizCategories(uint(quizID), quizData.Categories, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// เตรียม questions และ choices สำหรับการอัปเดต
	// 1. ลบคำถามและตัวเลือกทั้งหมดที่มีอยู่
	// 2. สร้างคำถามและตัวเลือกใหม่ตามข้อมูลที่ได้รับ

	// ดึงคำถามที่มีอยู่เดิม
	existingQuestions, err := h.quizService.GetQuestionsByQuizID(uint(quizID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// สร้าง map ของคำถามและตัวเลือกที่มีอยู่เดิม (ID เป็น key)
	existingQuestionMap := make(map[uint]models.Question)
	for _, q := range existingQuestions {
		existingQuestionMap[q.ID] = q
	}

	// ตรวจสอบว่าคำถามใดควรลบ
	keepQuestionIDs := make(map[uint]bool)
	for _, q := range quizData.Questions {
		if q.ID > 0 {
			keepQuestionIDs[q.ID] = true
		}
	}

	// ลบคำถามที่ไม่อยู่ในข้อมูลที่ได้รับ
	for _, q := range existingQuestions {
		if !keepQuestionIDs[q.ID] {
			if err := h.quizService.DeleteQuestion(q.ID, userID); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to delete question %d: %s", q.ID, err.Error()),
				})
			}
		}
	}

	// สร้างหรืออัปเดตคำถามและตัวเลือกตามข้อมูลที่ได้รับ
	for i, qData := range quizData.Questions {
		var question models.Question
		isNewQuestion := qData.ID == 0

		if isNewQuestion {
			// สร้างคำถามใหม่
			question = models.Question{
				QuizID: uint(quizID),
				Text:   qData.Text,
			}
		} else {
			// อัปเดตคำถามที่มีอยู่
			existingQuestion, exists := existingQuestionMap[qData.ID]
			if !exists {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Question with ID %d does not exist", qData.ID),
				})
			}
			question = existingQuestion
			question.Text = qData.Text
		}

		// อัปโหลดรูปภาพคำถาม (ถ้ามี)
		questionImageField := fmt.Sprintf("questionImage_%d", i)
		questionImage, err := c.FormFile(questionImageField)
		if err == nil && questionImage != nil {
			// ถ้ามีรูปภาพเดิม ให้ลบก่อน
			if question.ImageURL != "" {
				filePath, fileType, err := h.fileService.GetFilePath(question.ImageURL)
				if err == nil {
					_ = h.fileService.DeleteFile(filePath, fileType)
				}
			}

			// อัปโหลดรูปภาพใหม่
			questionImageURL, err := h.fileService.UploadFile(questionImage, "question")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to upload question image %d: %s", i, err.Error()),
				})
			}
			question.ImageURL = questionImageURL
		}

		// บันทึกหรืออัปเดตคำถาม
		if isNewQuestion {
			if err := h.quizService.CreateQuestion(&question, userID); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to create question %d: %s", i, err.Error()),
				})
			}
		} else {
			if err := h.quizService.UpdateQuestion(&question, userID); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to update question %d: %s", i, err.Error()),
				})
			}
		}

		// ดึงตัวเลือกที่มีอยู่เดิมของคำถามนี้
		var existingChoices []models.Choice
		if !isNewQuestion {
			existingChoices, err = h.quizService.GetChoicesByQuestionID(question.ID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to get choices for question %d: %s", question.ID, err.Error()),
				})
			}
		}

		existingChoiceMap := make(map[uint]models.Choice)
		for _, c := range existingChoices {
			existingChoiceMap[c.ID] = c
		}

		// ตรวจสอบว่าตัวเลือกใดควรลบ
		keepChoiceIDs := make(map[uint]bool)
		for _, c := range qData.Choices {
			if c.ID > 0 {
				keepChoiceIDs[c.ID] = true
			}
		}

		// ลบตัวเลือกที่ไม่อยู่ในข้อมูลที่ได้รับ
		for _, c := range existingChoices {
			if !keepChoiceIDs[c.ID] {
				if err := h.quizService.DeleteChoice(c.ID, userID); err != nil {
				}
			}
		}

		// สร้างหรืออัปเดตตัวเลือก
		for j, cData := range qData.Choices {
			var choice models.Choice
			isNewChoice := cData.ID == 0

			if isNewChoice {
				// สร้างตัวเลือกใหม่
				choice = models.Choice{
					QuestionID: question.ID,
					Text:       cData.Text,
					IsCorrect:  cData.IsCorrect,
				}
			} else {
				// อัปเดตตัวเลือกที่มีอยู่
				existingChoice, exists := existingChoiceMap[cData.ID]
				if !exists {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": fmt.Sprintf("Choice with ID %d does not exist", cData.ID),
					})
				}
				choice = existingChoice
				choice.Text = cData.Text
				choice.IsCorrect = cData.IsCorrect
			}

			// อัปโหลดรูปภาพตัวเลือก (ถ้ามี)
			choiceImageField := fmt.Sprintf("choiceImage_%d_%d", i, j)
			choiceImage, err := c.FormFile(choiceImageField)
			if err == nil && choiceImage != nil {
				// ถ้ามีรูปภาพเดิม ให้ลบก่อน
				if choice.ImageURL != "" {
					filePath, fileType, err := h.fileService.GetFilePath(choice.ImageURL)
					if err == nil {
						_ = h.fileService.DeleteFile(filePath, fileType)
					}
				}

				// อัปโหลดรูปภาพใหม่
				choiceImageURL, err := h.fileService.UploadFile(choiceImage, "choice")
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to upload choice image %d-%d: %s", i, j, err.Error()),
					})
				}
				choice.ImageURL = choiceImageURL
			}

			// บันทึกหรืออัปเดตตัวเลือก
			if isNewChoice {
				if err := h.quizService.CreateChoice(&choice, userID); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to create choice %d-%d: %s", i, j, err.Error()),
					})
				}
			} else {
				if err := h.quizService.UpdateChoice(&choice, userID); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to update choice %d-%d: %s", i, j, err.Error()),
					})
				}
			}
		}
	}
	log.Printf("QuizID: %d", quizID)
	return c.JSON(fiber.Map{
		"message": "Quiz updated successfully",
		"data": fiber.Map{
			"id": quizID,
		},
	})
}