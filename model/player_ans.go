package models

import (
	"time"
)

type PlayerAnswer struct {
	ID         uint        `gorm:"primaryKey"`
	SessionID  string      `gorm:"not null;index"` // string ธรรมดา
	Session    GameSession `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	QuizID     uint        `gorm:"not null;index"`
	Quiz       Quiz        `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	QuestionID uint        `gorm:"not null;index"`
	Question   Question    `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PlayerID   uint        `gorm:"not null;index"`
	Player     User        `gorm:"foreignKey:PlayerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ChoiceID   uint        `gorm:"not null;index"`
	Choice     Choice      `gorm:"foreignKey:ChoiceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	TimeSpent  float64     `gorm:"not null"`
	IsCorrect  bool        `gorm:"not null"`
	Points     uint        `gorm:"not null"`
	CreatedAt  time.Time
}
