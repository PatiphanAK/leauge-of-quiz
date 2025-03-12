package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/services"
	"github.com/patiphanak/league-of-quiz/utils"
)

// ChoiceHandler สำหรับจัดการ HTTP requests เกี่ยวกับตัวเลือก
type ChoiceHandler struct {
	choiceService *services.ChoiceService
	fileService   *services.FileService
}

// NewChoiceHandler สร้าง instance ใหม่ของ ChoiceHandler
func NewChoiceHandler(choiceService *services.ChoiceService, fileService *services.FileService) *ChoiceHandler {
	return &ChoiceHandler{
		choiceService: choiceService,
		fileService:   fileService,
	}
}

// GetChoicesByQuestionID ดึงตัวเลือกทั้งหมดของคำถาม
func (h *ChoiceHandler) GetChoicesByQuestionID(c *fiber.Ctx) error {
	// รับ ID จาก parameter
	questionID, statusCode, err := utils.ParseIDParam(c, "questionId")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงข้อมูลตัวเลือก
	choices, err := h.choiceService.GetChoicesByQuestionID(questionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": choices,
	})
}

// GetChoiceByID ดึงข้อมูลตัวเลือกจาก ID
func (h *ChoiceHandler) GetChoiceByID(c *fiber.Ctx) error {
	// รับ ID จาก parameter
	choiceID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงข้อมูลตัวเลือก
	choice, err := h.choiceService.GetChoiceByID(choiceID)
	if err != nil {
		// เช็คว่า error มาจากการไม่พบตัวเลือกหรือ Database Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Choice not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.JSON(fiber.Map{
		"data": choice,
	})
}

// CreateChoice สร้างตัวเลือกใหม่
func (h *ChoiceHandler) CreateChoice(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ตรวจสอบว่าเป็น multipart form หรือไม่
	if !strings.Contains(c.Get("Content-Type"), "multipart/form-data") {
		// รับข้อมูลจาก JSON body
		var choice models.Choice
		if err := c.BodyParser(&choice); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// ตรวจสอบข้อมูลที่จำเป็น
		if choice.Text == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Choice text is required",
			})
		}

		if choice.QuestionID == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Question ID is required",
			})
		}

		// สร้างตัวเลือก
		if err := h.choiceService.CreateChoice(&choice, nil, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Choice created successfully",
			"data": fiber.Map{
				"id": choice.ID,
			},
		})
	} else {
		// กรณีเป็น multipart form
		// ดึงข้อมูลจาก form
		text := c.FormValue("text")
		questionIDStr := c.FormValue("questionId")
		isCorrectStr := c.FormValue("isCorrect", "false")

		// ตรวจสอบข้อมูลที่จำเป็น
		if text == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Choice text is required",
			})
		}

		if questionIDStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Question ID is required",
			})
		}

		// แปลงค่า questionID และ isCorrect
		questionID, err := strconv.ParseUint(questionIDStr, 10, 32)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid Question ID",
			})
		}
		isCorrect := isCorrectStr == "true"

		// รับไฟล์รูปภาพ
		imageFile, _ := c.FormFile("image")

		// สร้าง choice object
		choice := &models.Choice{
			QuestionID: uint(questionID),
			Text:       text,
			IsCorrect:  isCorrect,
		}

		// สร้างตัวเลือกพร้อมรูปภาพ
		if err := h.choiceService.CreateChoice(choice, imageFile, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Choice created successfully",
			"data": fiber.Map{
				"id": choice.ID,
			},
		})
	}
}

// UpdateChoice อัปเดตข้อมูลตัวเลือก
func (h *ChoiceHandler) UpdateChoice(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ ID จาก parameter
	choiceID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ตรวจสอบว่าเป็น multipart form หรือไม่
	if !strings.Contains(c.Get("Content-Type"), "multipart/form-data") {
		// รับข้อมูลจาก JSON body
		var choiceData models.Choice
		if err := c.BodyParser(&choiceData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// ตรวจสอบข้อมูลที่จำเป็น
		if choiceData.Text == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Choice text is required",
			})
		}

		// กำหนด ID ให้กับตัวเลือก
		choiceData.ID = choiceID

		// อัปเดตข้อมูลตัวเลือก
		if err := h.choiceService.UpdateChoice(&choiceData, nil, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	} else {
		// กรณีเป็น multipart form
		// ดึงข้อมูลจาก form
		text := c.FormValue("text")
		isCorrectStr := c.FormValue("isCorrect", "")

		// ตรวจสอบข้อมูลที่จำเป็น
		if text == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Choice text is required",
			})
		}

		// ดึงข้อมูลตัวเลือกเดิม
		existingChoice, err := h.choiceService.GetChoiceByID(choiceID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Choice not found",
			})
		}

		// แปลงค่า isCorrect
		isCorrect := existingChoice.IsCorrect
		if isCorrectStr != "" {
			isCorrect = isCorrectStr == "true"
		}

		// รับไฟล์รูปภาพ
		imageFile, _ := c.FormFile("image")

		// สร้าง choice object สำหรับอัปเดต
		choice := &models.Choice{
			ID:         choiceID,
			QuestionID: existingChoice.QuestionID,
			Text:       text,
			IsCorrect:  isCorrect,
		}

		// อัปเดตข้อมูลตัวเลือกพร้อมรูปภาพ
		if err := h.choiceService.UpdateChoice(choice, imageFile, userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	// ดึงข้อมูลตัวเลือกที่อัปเดตแล้วและส่งกลับ
	updatedChoice, err := h.choiceService.GetChoiceByID(choiceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve updated choice",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Choice updated successfully",
		"data":    updatedChoice,
	})
}

// DeleteChoice ลบตัวเลือก
func (h *ChoiceHandler) DeleteChoice(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ ID จาก parameter
	choiceID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// ลบตัวเลือก
	if err := h.choiceService.DeleteChoice(choiceID, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Choice deleted successfully",
	})
}

// UploadChoiceImage อัปโหลดรูปภาพสำหรับตัวเลือก
func (h *ChoiceHandler) UploadChoiceImage(c *fiber.Ctx) error {
	// ตรวจสอบว่าผู้ใช้ล็อกอินแล้ว
	userID, statusCode, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับ ID จาก parameter
	choiceID, statusCode, err := utils.ParseIDParam(c, "id")
	if err != nil {
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	// รับไฟล์รูปภาพ
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded or invalid file",
		})
	}

	// ดึงข้อมูลตัวเลือกเดิม
	existingChoice, err := h.choiceService.GetChoiceByID(choiceID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Choice not found",
		})
	}

	// สร้าง choice object สำหรับอัปเดต (คงค่าเดิมยกเว้นรูปภาพ)
	choice := &models.Choice{
		ID:         choiceID,
		QuestionID: existingChoice.QuestionID,
		Text:       existingChoice.Text,
		IsCorrect:  existingChoice.IsCorrect,
	}

	// อัปโหลดรูปภาพ
	if err := h.choiceService.UpdateChoice(choice, file, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ดึงข้อมูลตัวเลือกที่อัปเดตแล้วและส่งกลับ
	updatedChoice, err := h.choiceService.GetChoiceByID(choiceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve updated choice",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Choice image uploaded successfully",
		"data": fiber.Map{
			"imageURL": updatedChoice.ImageURL,
		},
	})
}