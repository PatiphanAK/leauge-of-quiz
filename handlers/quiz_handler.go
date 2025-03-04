package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	model "github.com/patiphanak/league-of-quiz/model"
	validator "github.com/patiphanak/league-of-quiz/validator"
	"gorm.io/gorm"
)

// QuizHandler handles quiz-related endpoints
type QuizHandler struct {
	DB *gorm.DB
}

// NewQuizHandler creates a new quiz handler
func NewQuizHandler(db *gorm.DB) *QuizHandler {
	return &QuizHandler{DB: db}
}

// GetQuizzes returns a list of quizzes with optional filtering
func (h *QuizHandler) GetQuizzes(c *fiber.Ctx) error {
	// Parse query parameters
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	search := c.Query("search")
	categories := c.Query("categories")

	// Build the query
	query := h.DB.Model(&model.Quiz{})

	// Apply filters
	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	if categories != "" {
		categoryIDs := strings.Split(categories, ",")
		query = query.Joins("JOIN quiz_categories ON quiz_categories.quiz_id = quizzes.id").
			Where("quiz_categories.category_id IN ?", categoryIDs).
			Group("quizzes.id") // Prevent duplicates
	}

	// Fetch quizzes with pagination
	var quizzes []model.Quiz
	if err := query.Limit(limit).Offset(offset).Find(&quizzes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quizzes: " + err.Error(),
		})
	}

	return c.JSON(quizzes)
}

// GetQuizByID returns a quiz by its ID with questions and choices
func (h *QuizHandler) GetQuizByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	var quiz model.Quiz
	err = h.DB.Preload("Questions", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC") // Order questions
	}).Preload("Questions.Choices", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC") // Order choices
	}).First(&quiz, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Quiz not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quiz: " + err.Error(),
		})
	}

	return c.JSON(quiz)
}

// CreateQuiz creates a new quiz with questions and choices
func (h *QuizHandler) CreateQuiz(c *fiber.Ctx) error {
	var quiz model.Quiz
	if err := c.BodyParser(&quiz); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate the quiz
	errors := validator.ValidateQuiz(quiz)
	if len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errors,
		})
	}

	// Use a transaction to ensure data consistency
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Create the quiz
		if err := tx.Create(&quiz).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create quiz: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(quiz)
}

// UpdateQuiz updates an existing quiz
func (h *QuizHandler) UpdateQuiz(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// Check if quiz exists
	var quiz model.Quiz
	if err := h.DB.First(&quiz, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Quiz not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quiz: " + err.Error(),
		})
	}

	// Parse request body
	if err := c.BodyParser(&quiz); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	// Validate the quiz
	errors := validator.ValidateQuiz(quiz)
	if len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": errors,
		})
	}

	// Update the quiz
	if err := h.DB.Save(&quiz).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update quiz: " + err.Error(),
		})
	}

	return c.JSON(quiz)
}

// DeleteQuiz deletes a quiz and its associated data
func (h *QuizHandler) DeleteQuiz(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// Check if quiz exists
	var quiz model.Quiz
	if err := h.DB.First(&quiz, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Quiz not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quiz: " + err.Error(),
		})
	}

	// Use a transaction to delete the quiz and related data
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Delete related records first (cascade delete not used for explicit control)

		// Get questions
		var questions []model.Question
		if err := tx.Where("quiz_id = ?", id).Find(&questions).Error; err != nil {
			return err
		}

		// Delete choices for each question
		for _, question := range questions {
			if err := tx.Where("question_id = ?", question.ID).Delete(&model.Choice{}).Error; err != nil {
				return err
			}
		}

		// Delete questions
		if err := tx.Where("quiz_id = ?", id).Delete(&model.Question{}).Error; err != nil {
			return err
		}

		// Delete category associations
		if err := tx.Where("quiz_id = ?", id).Delete(&model.QuizCategory{}).Error; err != nil {
			return err
		}

		// Delete the quiz
		if err := tx.Delete(&quiz).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete quiz: " + err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
