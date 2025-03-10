package models

import "time"

type Quiz struct {
	ID          uint   `gorm:"primaryKey"` // ใช้ primaryKey แทน primary_key
	Title       string `gorm:"not null"`
	Description string `gorm:"not null"`
	TimeLimit   uint   `gorm:"not null;default:10"` // เพิ่ม default value
	IsPublished bool   `gorm:"default:false"`
	ImageURL    string `gorm:"default:null"`

	// สร้าง relation กับ User
	CreatorID  uint `gorm:"not null;index"` // เพิ่ม index
	Creator    User `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Questions  []Question `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Categories []Category `gorm:"many2many:quiz_categories;"`
}

type Question struct {
	ID       uint     `gorm:"primaryKey"`
	QuizID   uint     `gorm:"not null;index"`
	Quiz     Quiz     `gorm:"foreignKey:QuizID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // เพิ่ม foreignKey
	Text     string   `gorm:"not null"`
	ImageURL string   `gorm:"default:null"`
	Choices  []Choice `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Choice struct {
	ID         uint     `gorm:"primaryKey"`
	QuestionID uint     `gorm:"not null;index"`
	ImageURL   string   `gorm:"default:null"`
	Question   Question `gorm:"foreignKey:QuestionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"` // เพิ่ม foreignKey
	Text       string   `gorm:"not null"`
	IsCorrect  bool     `gorm:"not null"`
}

type Category struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;uniqueIndex"`
	Description string `gorm:"default:''"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Quizzes     []Quiz `gorm:"many2many:quiz_categories;"`
}

// QuizCategory represents the many-to-many relationship between quizzes and categories
type QuizCategory struct {
	QuizID     uint `gorm:"primaryKey"`
	CategoryID uint `gorm:"primaryKey"`
	CreatedAt  time.Time
}
