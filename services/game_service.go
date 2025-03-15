package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// GameService จัดการ business logic ของเกม
type GameService struct {
	repos            *repositories.Repositories // เปลี่ยนเป็นเก็บ repositories ทั้งหมด
	gameSessionRepo  *repositories.GameSessionRepository
	gamePlayerRepo   *repositories.GamePlayerRepository
	playerAnswerRepo *repositories.PlayerAnswerRepository
	choiceRepo       *repositories.ChoiceRepository
}

// NewGameService สร้าง GameService ใหม่
func NewGameService(
	repos *repositories.Repositories,
	gameSessionRepo *repositories.GameSessionRepository,
	gamePlayerRepo *repositories.GamePlayerRepository,
	playerAnswerRepo *repositories.PlayerAnswerRepository,
	choiceRepo *repositories.ChoiceRepository,
) *GameService {
	return &GameService{
		repos:            repos,
		gameSessionRepo:  gameSessionRepo,
		gamePlayerRepo:   gamePlayerRepo,
		playerAnswerRepo: playerAnswerRepo,
		choiceRepo:       choiceRepo,
	}
}

// CreateGameSession สร้าง session เกมใหม่ ด้วย transaction
func (s *GameService) CreateGameSession(hostID uint, quizID uint) (*models.GameSession, error) {
	// สร้าง ID สำหรับ session
	sessionID := uuid.New().String()

	// สร้าง GameSession
	now := time.Now()
	session := &models.GameSession{
		ID:        sessionID,
		QuizID:    quizID,
		HostID:    hostID,
		Status:    "lobby", // สถานะเริ่มต้นคือ lobby
		CreatedAt: now,
	}

	// ลงทะเบียนโฮสต์เป็นผู้เล่นด้วย
	hostPlayer := &models.GamePlayer{
		SessionID: sessionID,
		UserID:    hostID,
		Nickname:  "Host", // ตั้งชื่อเริ่มต้น สามารถเปลี่ยนได้ภายหลัง
		Score:     0,
		JoinedAt:  now,
	}

	// เริ่ม transaction
	tx := s.repos.BeginTx()

	// บันทึก session ในฐานข้อมูลใน transaction
	if err := tx.Create(session).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// บันทึกผู้เล่นโฮสต์ในฐานข้อมูลใน transaction
	if err := tx.Create(hostPlayer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// หากทุกอย่างเรียบร้อย commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return session, nil
}

// 3. ควรปรับปรุงฟังก์ชันอื่นๆ ที่ต้องการความเป็นอะตอมมิก เช่น SubmitAnswer

// SubmitAnswer บันทึกคำตอบของผู้เล่น ใช้ transaction
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
	existingAnswer, err := s.playerAnswerRepo.GetPlayerAnswerBySessionAndQuestion(sessionID, questionID, playerID)
	if err == nil && existingAnswer != nil {
		return nil, errors.New("ผู้เล่นได้ตอบคำถามนี้ไปแล้ว")
	}

	// ตรวจสอบว่าตัวเลือกนี้เป็นคำตอบที่ถูกต้องหรือไม่
	choice, err := s.choiceRepo.GetChoiceByID(choiceID)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบว่า choice นี้เป็นของ question นี้จริงๆ
	if choice.QuestionID != questionID {
		return nil, errors.New("ตัวเลือกนี้ไม่ได้อยู่ในคำถามที่ระบุ")
	}

	// คำนวณคะแนน
	isCorrect := choice.IsCorrect
	var points uint = 0
	if isCorrect {
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

	// เริ่ม transaction
	tx := s.repos.BeginTx()

	// บันทึกคำตอบในฐานข้อมูล
	if err := tx.Create(answer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ดึงข้อมูลผู้เล่นเพื่ออัพเดทคะแนน
	var player models.GamePlayer
	if err := tx.Where("session_id = ? AND user_id = ?", sessionID, playerID).First(&player).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// อัพเดทคะแนนผู้เล่น
	player.Score += points
	if err := tx.Save(&player).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// หากทุกอย่างเรียบร้อย commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return answer, nil
}

func (s *GameService) GetGameSession(sessionID string) (*models.GameSession, error) {
	return s.gameSessionRepo.GetGameSessionByID(sessionID)
}

// JoinGameSession ให้ผู้เล่นเข้าร่วม session (ปรับปรุงด้วย transaction)
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

	// ใช้ transaction
	tx := s.repos.BeginTx()
	if err := tx.Create(player).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return player, nil
}

// StartGameSession เริ่มเกม (ปรับปรุงด้วย transaction)
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

	// ใช้ transaction
	tx := s.repos.BeginTx()
	if err := tx.Save(session).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// EndGameSession จบเกม (ปรับปรุงด้วย transaction)
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

	// ใช้ transaction
	tx := s.repos.BeginTx()
	if err := tx.Save(session).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
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
