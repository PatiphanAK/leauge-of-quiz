package repositories

import (
	"errors"
	"log"

	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

type QuizRepository struct {
	db *gorm.DB
}

func (r *QuizRepository) GetDB() *gorm.DB {
	return r.db
}

// NewQuizRepository สร้าง instance ใหม่ของ QuizRepository
func NewQuizRepository(db *gorm.DB) *QuizRepository {
	log.Println("NewQuizRepository")
	return &QuizRepository{db: db}
}

// CreateQuiz สร้าง quiz ใหม่
func (r *QuizRepository) CreateQuiz(quiz *models.Quiz) error {
	return r.db.Create(quiz).Error
}

// GetQuizByID ดึงข้อมูล quiz จาก ID
func (r *QuizRepository) GetQuizByID(id uint) (*models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.Preload("Questions.Choices").Preload("Categories").First(&quiz, id).Error
	if err != nil {
		return nil, err
	}
	return &quiz, nil
}

// GetAllQuizzes ดึงข้อมูล quizzes ทั้งหมด
func (r *QuizRepository) GetAllQuizzes(page, limit int) ([]models.Quiz, int64, error) {
	var quizzes []models.Quiz
	var count int64

	offset := (page - 1) * limit

	// นับจำนวน quizzes ทั้งหมด
	r.db.Model(&models.Quiz{}).Count(&count)

	// ดึงข้อมูล quizzes
	err := r.db.Preload("Categories").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&quizzes).Error

	if err != nil {
		return nil, 0, err
	}

	return quizzes, count, nil
}

// GetPublishedQuizzes ดึงข้อมูล quizzes ที่เผยแพร่แล้ว
func (r *QuizRepository) GetPublishedQuizzes(page, limit int) ([]models.Quiz, int64, error) {
	var quizzes []models.Quiz
	var count int64

	offset := (page - 1) * limit

	// นับจำนวน quizzes ที่เผยแพร่
	r.db.Model(&models.Quiz{}).Where("is_published = ?", true).Count(&count)

	// ดึงข้อมูล quizzes
	err := r.db.Where("is_published = ?", true).
		Preload("Categories").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&quizzes).Error

	if err != nil {
		return nil, 0, err
	}

	return quizzes, count, nil
}

// GetQuizzesByCreator ดึงข้อมูล quizzes ของผู้สร้าง
func (r *QuizRepository) GetQuizzesByCreator(creatorID uint, page, limit int) ([]models.Quiz, int64, error) {
	var quizzes []models.Quiz
	var count int64

	offset := (page - 1) * limit

	// นับจำนวน quizzes ของผู้สร้าง
	r.db.Model(&models.Quiz{}).Where("creator_id = ?", creatorID).Count(&count)

	// ดึงข้อมูล quizzes
	err := r.db.Where("creator_id = ?", creatorID).
		Preload("Categories").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&quizzes).Error

	if err != nil {
		return nil, 0, err
	}

	return quizzes, count, nil
}

// UpdateQuiz อัปเดตข้อมูล quiz
func (r *QuizRepository) UpdateQuiz(quiz *models.Quiz) error {
	return r.db.Model(quiz).Updates(quiz).Error
}

// CheckQuizOwnership ตรวจสอบว่า user เป็นเจ้าของ quiz หรือไม่
func (r *QuizRepository) CheckQuizOwnership(quizID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Quiz{}).
		Where("id = ? AND creator_id = ?", quizID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteQuiz ลบ quiz
func (r *QuizRepository) DeleteQuiz(id uint) error {
	return r.db.Delete(&models.Quiz{}, id).Error
}

// UpdateQuizCategories อัปเดตหมวดหมู่ของ quiz
func (r *QuizRepository) UpdateQuizCategories(quizID uint, categoryIDs []uint) error {
	// เริ่ม transaction
	tx := r.db.Begin()

	// ลบ relations เดิมออก
	if err := tx.Exec("DELETE FROM quiz_categories WHERE quiz_id = ?", quizID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// เพิ่ม relations ใหม่
	for _, categoryID := range categoryIDs {
		// ตรวจสอบว่า category มีอยู่หรือไม่
		var category models.Category
		if err := tx.First(&category, categoryID).Error; err != nil {
			tx.Rollback()
			return errors.New("category not found")
		}

		// เพิ่ม relation
		if err := tx.Exec("INSERT INTO quiz_categories (quiz_id, category_id) VALUES (?, ?)", quizID, categoryID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

// Modified function in repositories/quiz_repo.go

func (r *QuizRepository) GetFilteredQuizzes(offset, limit int, isPublished string, search string, categories []uint) ([]models.Quiz, int64, error) {
	var quizzes []models.Quiz
	var count int64

	// สร้าง query base
	query := r.db.Model(&models.Quiz{})

	// ใช้ transaction หรือ subquery สำหรับการนับจำนวน
	countQuery := r.db.Model(&models.Quiz{})

	// เพิ่มเงื่อนไขการกรอง
	if isPublished == "true" {
		query = query.Where("is_published = ?", true)
		countQuery = countQuery.Where("is_published = ?", true)
	} else if isPublished == "false" {
		query = query.Where("is_published = ?", false)
		countQuery = countQuery.Where("is_published = ?", false)
	}

	// ค้นหาจากชื่อหรือคำอธิบาย
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
		countQuery = countQuery.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// กรองตามหมวดหมู่
	if len(categories) > 0 {
		// ใช้ joins หรือ subquery เพื่อกรองตามหมวดหมู่
		query = query.Joins("JOIN quiz_categories ON quizzes.id = quiz_categories.quiz_id").
			Where("quiz_categories.category_id IN ?", categories).
			Group("quizzes.id")

		countQuery = countQuery.Joins("JOIN quiz_categories ON quizzes.id = quiz_categories.quiz_id").
			Where("quiz_categories.category_id IN ?", categories).
			Group("quizzes.id")
	}

	// นับจำนวน quizzes
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูล quizzes with Questions and Choices
	err := query.
		Preload("Categories").
		Preload("Questions").         // Add preloading of Questions
		Preload("Questions.Choices"). // Add preloading of Choices
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&quizzes).Error

	if err != nil {
		return nil, 0, err
	}

	return quizzes, count, nil
}
