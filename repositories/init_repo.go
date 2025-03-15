package repositories

import (
	"log"

	"gorm.io/gorm"
)

type Repositories struct {
	DB           *gorm.DB
	Quiz         *QuizRepository
	Question     *QuestionRepository
	Choice       *ChoiceRepository
	GameSession  *GameSessionRepository
	GamePlayer   *GamePlayerRepository
	PlayerAnswer *PlayerAnswerRepository
}

func InitRepositories(db *gorm.DB) *Repositories {
	log.Println("Initializing all repositories")

	return &Repositories{
		Quiz:         NewQuizRepository(db),
		Question:     NewQuestionRepository(db),
		Choice:       NewChoiceRepository(db),
		GameSession:  NewGameSessionRepository(db),
		GamePlayer:   NewGamePlayerRepository(db),
		PlayerAnswer: NewPlayerAnswerRepository(db),
	}
}

func (r *Repositories) BeginTx() *gorm.DB {
	return r.DB.Begin()
}
