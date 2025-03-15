package models

import (
	"time"
)

type GameSession struct {
	ID         string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	QuizID     uint   `gorm:"not null;index"`
	Quiz       Quiz   `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	HostID     uint   `gorm:"not null;index"`
	Host       User   `gorm:"foreignKey:HostID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Status     string `gorm:"not null"`
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	Players    []GamePlayer `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type GamePlayer struct {
	ID        uint        `gorm:"primaryKey"`
	SessionID string      `gorm:"not null;index"`
	UserID    uint        `gorm:"not null;index;uniqueIndex:idx_session_user"`
	Session   GameSession `gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User      User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Nickname  string      `gorm:"not null"`
	Score     uint        `gorm:"default:0"`
	JoinedAt  time.Time
}
