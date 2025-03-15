package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

// PlayerAnswerRepository handles database operations for player answers
type PlayerAnswerRepository struct {
	db *gorm.DB
}

// NewPlayerAnswerRepository creates a new player answer repository
func NewPlayerAnswerRepository(db *gorm.DB) *PlayerAnswerRepository {
	return &PlayerAnswerRepository{
		db: db,
	}
}

// CreatePlayerAnswer creates a new player answer
func (r *PlayerAnswerRepository) CreatePlayerAnswer(answer *models.PlayerAnswer) error {
	return r.db.Create(answer).Error
}

// GetPlayerAnswersBySessionID gets all answers for a session
func (r *PlayerAnswerRepository) GetPlayerAnswersBySessionID(sessionID string) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("session_id = ?", sessionID).Find(&answers).Error
	return answers, err
}

// GetPlayerAnswersByQuestionID gets all answers for a question
func (r *PlayerAnswerRepository) GetPlayerAnswersByQuestionID(questionID uint) ([]models.PlayerAnswer, error) {
	var answers []models.PlayerAnswer
	err := r.db.Where("question_id = ?", questionID).Find(&answers).Error
	return answers, err
}

// GetPlayerAnswerBySessionAndQuestion gets a player's answer for a specific question in a session
func (r *PlayerAnswerRepository) GetPlayerAnswerBySessionAndQuestion(sessionID string, questionID uint, playerID uint) (*models.PlayerAnswer, error) {
	var answer models.PlayerAnswer
	err := r.db.Where("session_id = ? AND question_id = ? AND player_id = ?", sessionID, questionID, playerID).First(&answer).Error
	if err != nil {
		return nil, err
	}
	return &answer, nil
}
