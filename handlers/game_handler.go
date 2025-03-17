package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/services"
)

// GameHandler จัดการ HTTP requests สำหรับเกม
type GameHandler struct {
	gameService *services.GameService
}

// NewGameHandler สร้าง GameHandler ใหม่
func NewGameHandler(gameService *services.GameService) *GameHandler {
	if gameService == nil {
		log.Fatal("gameService cannot be nil")
	}
	return &GameHandler{
		gameService: gameService,
	}
}

// CreateGameSessionRequest คือ request body สำหรับการสร้างเกม
type CreateGameSessionRequest struct {
	QuizID uint `json:"quizId"`
}

// CreateGameSession สร้างเกมใหม่
func (h *GameHandler) CreateGameSession(c *fiber.Ctx) error {

	// Validation: Check if gameService is initialized
	if h.gameService == nil {
		log.Println("Error: gameService is nil")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Game service not initialized",
		})
	}

	var req CreateGameSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validation: Check if QuizID is valid
	if req.QuizID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid quiz ID",
		})
	}

	// ดึง userID จาก context (ที่ตั้งโดย middleware)
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	log.Println("userID", userID)
	log.Println("req.QuizID", req.QuizID)

	session, err := h.gameService.CreateGameSession(userID, req.QuizID)
	if err != nil {
		log.Printf("Error creating game session: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	log.Println("session", session)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Game session created successfully",
		"session": session,
	})
}

// GetGameSessions ดึงเกมทั้งหมดที่กำลังรออยู่
func (h *GameHandler) GetGameSessions(c *fiber.Ctx) error {
	// ดึงเฉพาะเกมที่ผู้ใช้เป็นโฮสต์หรือไม่
	hostOnly := c.QueryBool("hostOnly", false)

	var sessions interface{}
	var err error

	if hostOnly {
		// ดึง userID จาก context
		userID, ok := c.Locals("userID").(uint)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not authenticated",
			})
		}

		// ดึงเกมที่ผู้ใช้เป็นโฮสต์
		sessions, err = h.gameService.GetSessionsByHostID(userID)
	} else {
		// ดึงเกมที่กำลังรออยู่ทั้งหมด
		sessions, err = h.gameService.GetActiveGameSessions()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}

// GetGameSessionDetail ดึงรายละเอียดของเกม
func (h *GameHandler) GetGameSessionDetail(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	session, err := h.gameService.GetGameSession(sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game session not found",
		})
	}

	// ดึงรายชื่อผู้เล่น
	players, err := h.gameService.GetPlayersBySessionID(sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get players",
		})
	}

	return c.JSON(fiber.Map{
		"session": session,
		"players": players,
	})
}

// JoinGameSessionRequest คือ request body สำหรับการเข้าร่วมเกม
type JoinGameSessionRequest struct {
	Nickname string `json:"nickname"`
}

// JoinGameSession เข้าร่วมเกม
func (h *GameHandler) JoinGameSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	var req JoinGameSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ดึง userID จาก context
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// ใช้ชื่อเล่นจาก request หรือสร้างชื่อเล่นจากเวลาปัจจุบัน
	nickname := req.Nickname
	if nickname == "" {
		nickname = "Player_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	player, err := h.gameService.JoinGameSession(sessionID, userID, nickname)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Joined game session successfully",
		"player":  player,
	})
}

// StartGameSession เริ่มเกม
func (h *GameHandler) StartGameSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	// ดึง userID จาก context
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	err := h.gameService.StartGameSession(sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Game started successfully",
	})
}

// SubmitAnswerRequest คือ request body สำหรับการส่งคำตอบ
type SubmitAnswerRequest struct {
	QuestionID uint    `json:"questionId"`
	ChoiceID   uint    `json:"choiceId"`
	TimeSpent  float64 `json:"timeSpent"`
}

// SubmitAnswer ส่งคำตอบ
func (h *GameHandler) SubmitAnswer(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	var req SubmitAnswerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// ดึง userID จาก context
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	answer, err := h.gameService.SubmitAnswer(sessionID, userID, req.QuestionID, req.ChoiceID, req.TimeSpent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Answer submitted successfully",
		"answer":  answer,
	})
}

// EndGameSession จบเกม
func (h *GameHandler) EndGameSession(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	// ดึง userID จาก context
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	session, err := h.gameService.EndGameSession(sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Game ended successfully",
		"session": session,
	})
}

// GetGameResults ดึงผลลัพธ์ของเกม
func (h *GameHandler) GetGameResults(c *fiber.Ctx) error {
	sessionID := c.Params("id")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Session ID is required",
		})
	}

	// ดึงข้อมูลเกม
	session, err := h.gameService.GetGameSession(sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Game session not found",
		})
	}

	// ดึงรายชื่อผู้เล่น
	players, err := h.gameService.GetPlayersBySessionID(sessionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get players",
		})
	}

	return c.JSON(fiber.Map{
		"session": session,
		"players": players,
	})
}
