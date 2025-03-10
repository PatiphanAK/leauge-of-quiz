package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	GoogleID    string `gorm:"unique"`
	Email       string
	DisplayName string
	PictureURL  string
	Quiz        []Quiz `gorm:"foreignKey:CreatorID" json:"-"`
}
