package models

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB เริ่มต้นการเชื่อมต่อกับฐานข้อมูล
func InitDB() {
	// ตั้งค่า Database Connection
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file, using environment variables")
	}

	// อ่านค่า Connection String จาก .env
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// เชื่อมต่อ Database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	fmt.Println("Database connected successfully")
}

// MigrateDB สร้างหรืออัปเดตตารางในฐานข้อมูล
func MigrateDB() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	err := DB.AutoMigrate(
		&User{},
		&Quiz{},
		&Question{},
		&Choice{},
		&PlayerAnswer{},
		&GamePlayer{},
		&GameSession{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}
