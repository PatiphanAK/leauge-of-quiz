package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/patiphanak/league-of-quiz/dto"
	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/services"
	"github.com/patiphanak/league-of-quiz/utils"
)

// QuestionHandler สำหรับการจัดการ API ที่เกี่ยวข้องกับคำถาม
type QuestionHandler struct {
	questionService *services.QuestionService
	fileService     *services.FileService
	choiceService   *services.ChoiceService
}

// NewQuestionHandler สร้าง instance ใหม่ของ QuestionHandler
func NewQuestionHandler(
	questionService *services.QuestionService, 
	fileService *services.FileService,
	choiceService *services.ChoiceService,
	) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
		fileService:     fileService,
		choiceService:  choiceService,
	}
}

// GetQuestionsByQuizID ดึงคำถามทั้งหมดของ quiz
func (h *QuestionHandler) GetQuestionsByQuizID(c *fiber.Ctx) error {
	// รับ quizID จาก parameter
	quizID, statusCode, err := utils.ParseIDParam(c, "quizId")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงคำถามทั้งหมดของ quiz
	questions, err := h.questionService.GetQuestionsByQuizID(quizID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": questions})
}

// GetQuestionByID ดึงข้อมูลคำถามจาก ID
func (h *QuestionHandler) GetQuestionByID(c *fiber.Ctx) error {
	// รับ questionID จาก parameter
	questionID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงข้อมูลคำถาม
	question, err := h.questionService.GetQuestionByID(questionID)
	if err != nil {
		// เช็คว่า error มาจากการไม่พบ Question หรือ Database Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Question not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{"data": question})
}

// CreateQuestion สร้างคำถามใหม่
func (h *QuestionHandler) CreateQuestion(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}
	// รับข้อมูลจาก multipart form
	quizIDStr := c.FormValue("quizId")
	text := c.FormValue("text")

	// ตรวจสอบข้อมูลที่จำเป็น
	log.Print("quizIDStr: ", quizIDStr)
	if quizIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quiz ID is required"})
	}
	if text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question text is required"})
	}

	// แปลง quizID เป็น uint
	quizID, err := strconv.ParseUint(quizIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Quiz ID"})
	}

	// รับไฟล์รูปภาพ (ถ้ามี)
	imageFile, _ := c.FormFile("image")

	// สร้างคำถาม
	question := &models.Question{
		QuizID: uint(quizID),
		Text:   text,
	}

	// บันทึกคำถามพร้อมรูปภาพ
	if err := h.questionService.CreateQuestion(question, imageFile, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	choices := c.FormValue("choices")
	log.Print("choices: ", choices)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Question created successfully",
		"data": fiber.Map{
			"id": question.ID,
		},
	})
}

// UpdateQuestion อัปเดตข้อมูลคำถาม
func (h *QuestionHandler) UpdateQuestion(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ questionID จาก parameter
	questionID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงข้อมูลคำถามเดิม
	existingQuestion, err := h.questionService.GetQuestionByID(questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Question not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// รับข้อมูลจาก multipart form
	text := c.FormValue("text", existingQuestion.Text)

	// ตรวจสอบข้อมูลที่จำเป็น
	if text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question text is required"})
	}

	// รับไฟล์รูปภาพ (ถ้ามี)
	imageFile, _ := c.FormFile("image")

	// อัปเดตข้อมูลคำถาม
	question := &models.Question{
		ID:     questionID,
		QuizID: existingQuestion.QuizID, // ไม่อนุญาตให้เปลี่ยน QuizID
		Text:   text,
	}

	// บันทึกข้อมูลคำถามที่อัปเดต
	if err := h.questionService.UpdateQuestion(question, imageFile, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Question updated successfully",
		"data": fiber.Map{
			"id": question.ID,
		},
	})
}

func (h *QuestionHandler) PatchQuestion(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ questionID จาก parameter
	questionID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับข้อมูล JSON จาก form
	questionDataStr := c.FormValue("questionData")
	if questionDataStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question data is required"})
	}

	// แปลง JSON เป็น struct
	var formData dto.QuestionFormData
	if err := json.Unmarshal([]byte(questionDataStr), &formData); err != nil {
		log.Printf("Error parsing question data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid question data format"})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if formData.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question text is required"})
	}

	// รับไฟล์รูปภาพคำถาม (ถ้ามี)
	questionImage, _ := c.FormFile("image")

	// รับไฟล์รูปภาพตัวเลือก
	choiceImages := make(map[int]*multipart.FileHeader)
	for i := range formData.Choices {
		choiceImage, _ := c.FormFile(fmt.Sprintf("choices[%d][image]", i))
		if choiceImage != nil {
			choiceImages[i] = choiceImage
		}
	}

	// เรียกใช้ service เพื่ออัปเดตคำถามและตัวเลือกทั้งหมดในครั้งเดียว
	err = h.questionService.UpdateQuestionWithChoices(
		questionID,
		formData.Text,
		formData.Choices,
		questionImage,
		choiceImages,
		userID,
	)

	if err != nil {
		log.Printf("Error updating question with choices: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Question updated successfully",
		"data": fiber.Map{
			"id": questionID,
		},
	})
}

// DeleteQuestion ลบคำถาม
func (h *QuestionHandler) DeleteQuestion(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ questionID จาก parameter
	questionID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ลบคำถาม
	if err := h.questionService.DeleteQuestion(questionID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Question deleted successfully",
	})
}