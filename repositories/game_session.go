package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

// GameSessionRepository handles database operations for game sessions
type GameSessionRepository struct {
	db *gorm.DB
}

// NewGameSessionRepository creates a new game session repository
func NewGameSessionRepository(db *gorm.DB) *GameSessionRepository {
	return &GameSessionRepository{
		db: db,
	}
}

// CreateGameSession creates a new game session
func (r *GameSessionRepository) CreateGameSession(session *models.GameSession) error {
	return r.db.Create(session).Error
}

// GetGameSessionByID gets a game session by ID
func (r *GameSessionRepository) GetGameSessionByID(id string) (*models.GameSession, error) {
	var session models.GameSession
	err := r.db.Preload("Quiz").Preload("Host").Preload("Players").Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateGameSession updates a game session
func (r *GameSessionRepository) UpdateGameSession(session *models.GameSession) error {
	return r.db.Save(session).Error
}

// DeleteGameSession deletes a game session
func (r *GameSessionRepository) DeleteGameSession(id string) error {
	return r.db.Delete(&models.GameSession{}, "id = ?", id).Error
}

// GetSessionsByHostID gets all sessions hosted by a user
func (r *GameSessionRepository) GetSessionsByHostID(hostID uint) ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.Where("host_id = ?", hostID).Find(&sessions).Error
	return sessions, err
}

// GetActiveGameSessions gets all active game sessions (in lobby state)
func (r *GameSessionRepository) GetActiveGameSessions() ([]models.GameSession, error) {
	var sessions []models.GameSession
	err := r.db.Where("status = ?", "lobby").Preload("Quiz").Preload("Host").Find(&sessions).Error
	return sessions, err
}
