package database

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

func AutoMigration(db *gorm.DB) {
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Quiz{})
	db.AutoMigrate(&models.Question{})
	db.AutoMigrate(&models.PlayerAnswer{})
	db.AutoMigrate(&models.Choice{})
	db.AutoMigrate(&models.GameSession{})
	db.AutoMigrate(&models.GamePlayer{})
}
