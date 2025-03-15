package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"

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
		choiceService:   choiceService,
	}
}

func getChoiceImages(c *fiber.Ctx, numChoices int) map[int]*multipart.FileHeader {
	choiceImages := make(map[int]*multipart.FileHeader)
	for i := 0; i < numChoices; i++ {
		choiceImage, err := c.FormFile(fmt.Sprintf("choices[%d][image]", i))
		if err == nil && choiceImage != nil {
			choiceImages[i] = choiceImage
		}
	}
	return choiceImages
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

	// รับข้อมูล JSON จาก form
	questionDataStr := c.FormValue("questionData")
	if questionDataStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question data is required"})
	}

	var formData dto.QuestionFormData
	if err := json.Unmarshal([]byte(questionDataStr), &formData); err != nil {
		log.Printf("Error parsing question data: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid question data format"})
	}

	if formData.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Question text is required"})
	}

	questionImage, _ := c.FormFile("image")

	choiceImages := getChoiceImages(c, len(formData.Choices))

	question := &models.Question{
		QuizID: uint(formData.QuizID),
		Text:   formData.Text,
	}

	// เรียกใช้ฟังก์ชันโดยส่ง question แทน
	questionID, err := h.questionService.CreateQuestionWithChoices(
		question,
		formData.Choices,
		questionImage,
		choiceImages,
		userID,
	)

	if err != nil {
		log.Printf("Error creating question with choices: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Question created successfully",
		"data": fiber.Map{
			"id": questionID,
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
