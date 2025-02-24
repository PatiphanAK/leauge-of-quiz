package models

import "time"

type Quiz struct {
	ID          uint   `gorm:"primaryKey"` // ใช้ primaryKey แทน primary_key
	Title       string `gorm:"not null"`
	Description string `gorm:"not null"`

	// สร้าง relation กับ User
	CreatorID uint `gorm:"not null;index"` // เพิ่ม index
	Creator   User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	IsPublished bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Questions   []Question `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Question struct {
	ID      uint     `gorm:"primaryKey"`
	QuizID  uint     `gorm:"not null;index"`
	Quiz    Quiz     `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // เพิ่ม foreignKey
	Text    string   `gorm:"not null"`
	Choices []Choice `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Choice struct {
	ID         uint     `gorm:"primaryKey"`
	QuestionID uint     `gorm:"not null;index"`
	Question   Question `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"` // เพิ่ม foreignKey
	Text       string   `gorm:"not null"`
	IsCorrect  bool     `gorm:"not null"`
}
