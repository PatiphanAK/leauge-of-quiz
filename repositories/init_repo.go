package repositories

import (
	"log"

	"gorm.io/gorm"
)

type Repositories struct {
	Quiz     *QuizRepository
	Question *QuestionRepository
	Choice   *ChoiceRepository
}

func InitRepositories(db *gorm.DB) *Repositories {
	log.Println("Initializing all repositories")

	return &Repositories{
		Quiz:     NewQuizRepository(db),
		Question: NewQuestionRepository(db),
		Choice:   NewChoiceRepository(db),
	}
}
