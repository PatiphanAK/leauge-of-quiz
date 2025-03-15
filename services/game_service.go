package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// GameService คือส่วนที่จัดการ business logic ของเกม
type GameService struct {
	gameSessionRepo  *repositories.GameSessionRepository
	gamePlayerRepo   *repositories.GamePlayerRepository
	playerAnswerRepo *repositories.PlayerAnswerRepository
	choiceRepo       *repositories.ChoiceRepository
	// เพิ่ม repository อื่นๆ ตามความจำเป็น
}

// NewGameService สร้าง GameService ใหม่
func NewGameService(
	gameSessionRepo *repositories.GameSessionRepository,
	gamePlayerRepo *repositories.GamePlayerRepository,
	playerAnswerRepo *repositories.PlayerAnswerRepository,
	choiceRepo *repositories.ChoiceRepository,
) *GameService {
	return &GameService{
		gameSessionRepo:  gameSessionRepo,
		gamePlayerRepo:   gamePlayerRepo,
		playerAnswerRepo: playerAnswerRepo,
		choiceRepo:       choiceRepo,
	}
}

// CreateGameSession สร้าง session เกมใหม่
func (s *GameService) CreateGameSession(hostID uint, quizID uint) (*models.GameSession, error) {
	// สร้าง ID สำหรับ session
	sessionID := uuid.New()

	// สร้าง GameSession
	now := time.Now()
	session := &models.GameSession{
		ID:        sessionID,
		QuizID:    quizID,
		HostID:    hostID,
		Status:    "lobby", // สถานะเริ่มต้นคือ lobby
		CreatedAt: now,
	}

	// บันทึกลงฐานข้อมูล
	err := s.gameSessionRepo.CreateGameSession(session)
	if err != nil {
		return nil, err
	}

	// ลงทะเบียนโฮสต์เป็นผู้เล่นด้วย
	hostPlayer := &models.GamePlayer{
		SessionID: sessionID.String(),
		UserID:    hostID,
		Nickname:  "Host", // ตั้งชื่อเริ่มต้น สามารถเปลี่ยนได้ภายหลัง
		Score:     0,
		JoinedAt:  now,
	}

	err = s.gamePlayerRepo.CreateGamePlayer(hostPlayer)
	if err != nil {
		// ถ้าสร้างผู้เล่นไม่สำเร็จ ให้ลบ session
		s.gameSessionRepo.DeleteGameSession(sessionID.String())
		return nil, err
	}

	return session, nil
}

// GetGameSession ดึงข้อมูล game session จาก ID
func (s *GameService) GetGameSession(sessionID string) (*models.GameSession, error) {
	return s.gameSessionRepo.GetGameSessionByID(sessionID)
}

// JoinGameSession ให้ผู้เล่นเข้าร่วม session
func (s *GameService) JoinGameSession(sessionID string, userID uint, nickname string) (*models.GamePlayer, error) {
	// ตรวจสอบว่า session มีอยู่จริงและยังอยู่ในสถานะ lobby
	session, err := s.gameSessionRepo.GetGameSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != "lobby" {
		return nil, errors.New("ไม่สามารถเข้าร่วมได้: เกมได้เริ่มต้นหรือจบไปแล้ว")
	}

	// ตรวจสอบว่าผู้เล่นอยู่ใน session นี้แล้วหรือไม่
	existingPlayer, err := s.gamePlayerRepo.GetPlayerBySessionAndUserID(sessionID, userID)
	if err == nil && existingPlayer != nil {
		// ผู้เล่นอยู่ใน session นี้แล้ว
		return existingPlayer, nil
	}

	// สร้างผู้เล่นใหม่
	player := &models.GamePlayer{
		SessionID: sessionID,
		UserID:    userID,
		Nickname:  nickname,
		Score:     0,
		JoinedAt:  time.Now(),
	}

	err = s.gamePlayerRepo.CreateGamePlayer(player)
	if err != nil {
		return nil, err
	}

	return player, nil
}

// StartGameSession เริ่มเกม
func (s *GameService) StartGameSession(sessionID string, hostID uint) error {
	// ตรวจสอบว่าผู้ร้องขอคือโฮสต์
	session, err := s.gameSessionRepo.GetGameSessionByID(sessionID)
	if err != nil {
		return err
	}

	if session.HostID != hostID {
		return errors.New("เฉพาะโฮสต์เท่านั้นที่สามารถเริ่มเกมได้")
	}

	if session.Status != "lobby" {
		return errors.New("เกมได้เริ่มต้นหรือจบไปแล้ว")
	}

	// อัพเดทสถานะเป็น "in_progress"
	startTime := time.Now()
	session.Status = "in_progress"
	session.StartedAt = &startTime

	return s.gameSessionRepo.UpdateGameSession(session)
}

// SubmitAnswer บันทึกคำตอบของผู้เล่น
func (s *GameService) SubmitAnswer(sessionID string, playerID uint, questionID uint, choiceID uint, timeSpent float64) (*models.PlayerAnswer, error) {
	// ตรวจสอบว่า session อยู่ในสถานะ in_progress
	session, err := s.gameSessionRepo.GetGameSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != "in_progress" {
		return nil, errors.New("ไม่สามารถส่งคำตอบได้: เกมไม่ได้อยู่ในสถานะกำลังเล่น")
	}

	// ตรวจสอบว่าผู้เล่นได้ตอบคำถามนี้ไปแล้วหรือไม่
	// ต้องสร้าง repository method ใหม่หรือปรับปรุงจากที่มีอยู่

	// ตรวจสอบว่าตัวเลือกนี้เป็นคำตอบที่ถูกต้องหรือไม่
	choice, err := s.choiceRepo.GetChoiceByID(choiceID)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า choice นี้เป็นของ question นี้จริงๆ
	if choice.QuestionID != questionID {
		return nil, errors.New("ตัวเลือกนี้ไม่ได้อยู่ในคำถามที่ระบุ")
	}

	// คำนวณคะแนน - ตัวอย่างง่ายๆ
	isCorrect := choice.IsCorrect
	var points uint = 0
	if isCorrect {
		// คำนวณคะแนนตามเวลาที่ใช้ - ยิ่งเร็วยิ่งได้คะแนนมาก
		// นี่เป็นเพียงตัวอย่าง คุณอาจจะมีการคำนวณที่ซับซ้อนกว่านี้
		if timeSpent <= 5 {
			points = 100
		} else if timeSpent <= 10 {
			points = 75
		} else if timeSpent <= 15 {
			points = 50
		} else {
			points = 25
		}
	}

	// บันทึกคำตอบ
	answer := &models.PlayerAnswer{
		SessionID:  sessionID,
		QuizID:     session.QuizID,
		QuestionID: questionID,
		PlayerID:   playerID,
		ChoiceID:   choiceID,
		TimeSpent:  timeSpent,
		IsCorrect:  isCorrect,
		Points:     points,
		CreatedAt:  time.Now(),
	}

	err = s.playerAnswerRepo.CreatePlayerAnswer(answer)
	if err != nil {
		return nil, err
	}

	// อัพเดทคะแนนผู้เล่น
	player, err := s.gamePlayerRepo.GetPlayerBySessionAndUserID(sessionID, playerID)
	if err != nil {
		return answer, nil // ยังคืนคำตอบไปให้ แม้จะไม่สามารถอัพเดทคะแนนได้
	}

	player.Score += points
	err = s.gamePlayerRepo.UpdateGamePlayer(player)
	if err != nil {
		return answer, nil // ยังคืนคำตอบไปให้ แม้จะไม่สามารถอัพเดทคะแนนได้
	}

	return answer, nil
}

// EndGameSession จบเกม
func (s *GameService) EndGameSession(sessionID string, hostID uint) (*models.GameSession, error) {
	// ตรวจสอบว่าผู้ร้องขอคือโฮสต์
	session, err := s.gameSessionRepo.GetGameSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session.HostID != hostID {
		return nil, errors.New("เฉพาะโฮสต์เท่านั้นที่สามารถจบเกมได้")
	}

	if session.Status != "in_progress" {
		return nil, errors.New("เกมไม่ได้อยู่ในสถานะกำลังเล่น")
	}

	// อัพเดทสถานะเป็น "completed"
	finishedTime := time.Now()
	session.Status = "completed"
	session.FinishedAt = &finishedTime

	// อัพเดท session
	err = s.gameSessionRepo.UpdateGameSession(session)
	if err != nil {
		return nil, err
	}

	// ดึงข้อมูล session ล่าสุดพร้อมข้อมูลเพิ่มเติม
	return s.gameSessionRepo.GetGameSessionByID(sessionID)
}

// GetActiveGameSessions ดึงเกมที่กำลังรออยู่ (สถานะ lobby)
func (s *GameService) GetActiveGameSessions() ([]models.GameSession, error) {
	return s.gameSessionRepo.GetActiveGameSessions()
}

// GetSessionsByHostID ดึงเกมทั้งหมดที่ผู้ใช้เป็นโฮสต์
func (s *GameService) GetSessionsByHostID(hostID uint) ([]models.GameSession, error) {
	return s.gameSessionRepo.GetSessionsByHostID(hostID)
}

// GetPlayersBySessionID ดึงรายชื่อผู้เล่นทั้งหมดในเกม
func (s *GameService) GetPlayersBySessionID(sessionID string) ([]models.GamePlayer, error) {
	return s.gamePlayerRepo.GetPlayersBySessionID(sessionID)
}
