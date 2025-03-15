package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

type GamePlayerRepository struct {
	db *gorm.DB
}

// NewGamePlayerRepository creates a new game player repository
func NewGamePlayerRepository(db *gorm.DB) *GamePlayerRepository {
	return &GamePlayerRepository{
		db: db,
	}
}

// CreateGamePlayer creates a new game player
func (r *GamePlayerRepository) CreateGamePlayer(player *models.GamePlayer) error {
	return r.db.Create(player).Error
}

// GetGamePlayerByID gets a game player by ID
func (r *GamePlayerRepository) GetGamePlayerByID(id uint) (*models.GamePlayer, error) {
	var player models.GamePlayer
	err := r.db.First(&player, id).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// GetPlayerBySessionAndUserID gets a player by session ID and user ID
func (r *GamePlayerRepository) GetPlayerBySessionAndUserID(sessionID string, userID uint) (*models.GamePlayer, error) {
	var player models.GamePlayer
	err := r.db.Where("session_id = ? AND user_id = ?", sessionID, userID).First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

// UpdateGamePlayer updates a game player
func (r *GamePlayerRepository) UpdateGamePlayer(player *models.GamePlayer) error {
	return r.db.Save(player).Error
}

// GetPlayersBySessionID gets all players in a session
func (r *GamePlayerRepository) GetPlayersBySessionID(sessionID string) ([]models.GamePlayer, error) {
	var players []models.GamePlayer
	err := r.db.Where("session_id = ?", sessionID).Find(&players).Error
	return players, err
}
