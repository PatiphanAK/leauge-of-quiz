package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

// PlayerAnswerRepository จัดการการเข้าถึงข้อมูลคำตอบของผู้เล่น
type PlayerAnswerRepository struct {
	db *gorm.DB
}

// NewPlayerAnswerRepository สร้าง PlayerAnswerRepository ใหม่
func NewPlayerAnswerRepository(db *gorm.DB) *PlayerAnswerRepository {
	return &PlayerAnswerRepository{
		db: db,
	}
}

// CreatePlayerAnswer บันทึกคำตอบของผู้เล่น
func (r *PlayerAnswerRepository) CreatePlayerAnswer(answer *models.PlayerAnswer) error {
	return r.db.Create(answer).Error
}

// GetPlayerAnswersBySessionID ดึงคำตอบทั้งหมดในเกม
func (r *PlayerAnswerRepository) GetPlayerAnswersBySessionID(sessionID string) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("session_id = ?", sessionID).Find(&answers).Error
	return answers, err
}

// GetPlayerAnswersByQuestionID ดึงคำตอบทั้งหมดของคำถาม
func (r *PlayerAnswerRepository) GetPlayerAnswersByQuestionID(questionID uint) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("question_id = ?", questionID).Find(&answers).Error
	return answers, err
}

// GetPlayerAnswerBySessionAndQuestion ดึงคำตอบของผู้เล่นสำหรับคำถามเฉพาะในเกม
func (r *PlayerAnswerRepository) GetPlayerAnswerBySessionAndQuestion(sessionID string, questionID uint, playerID uint) (*models.PlayerAnswer, error) {
	var answer models.PlayerAnswer
	err := r.db.Where("session_id = ? AND question_id = ? AND player_id = ?", sessionID, questionID, playerID).First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

// GetPlayerAnswersByPlayerID ดึงคำตอบทั้งหมดของผู้เล่น
func (r *PlayerAnswerRepository) GetPlayerAnswersByPlayerID(playerID uint) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("player_id = ?", playerID).Find(&answers).Error
	return answers, err
}

// GetPlayerAnswersBySessionAndPlayerID ดึงคำตอบทั้งหมดของผู้เล่นในเกม
func (r *PlayerAnswerRepository) GetPlayerAnswersBySessionAndPlayerID(sessionID string, playerID uint) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("session_id = ? AND player_id = ?", sessionID, playerID).Find(&answers).Error
	return answers, err
}
