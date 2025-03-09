package repositories

import (
	"log"

	"gorm.io/gorm"
)

// Repositories holds all repository instances
type Repositories struct {
	Quiz *QuizRepository
}

// InitRepositories initializes all repositories at once
func InitRepositories(db *gorm.DB) *Repositories {
	log.Println("Initializing all repositories")

	return &Repositories{
		Quiz: NewQuizRepository(db),
	}
}
