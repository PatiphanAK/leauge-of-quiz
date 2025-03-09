package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

type QuestionRepository struct {
	db *gorm.DB
}

// NewQuestionRepository สร้าง instance ใหม่ของ QuestionRepository
func NewQuestionRepository(db *gorm.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// CreateQuestion สร้างคำถามใหม่
func (r *QuestionRepository) CreateQuestion(question *models.Question) error {
	return r.db.Create(question).Error
}

// GetQuestionByID ดึงข้อมูลคำถามจาก ID
func (r *QuestionRepository) GetQuestionByID(id uint) (*models.Question, error) {
	var question models.Question
	err := r.db.Preload("Choices").First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetQuestionsByQuizID ดึงคำถามทั้งหมดของ quiz
func (r *QuestionRepository) GetQuestionsByQuizID(quizID uint) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.Where("quiz_id = ?", quizID).Preload("Choices").Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

// UpdateQuestion อัปเดตข้อมูลคำถาม
func (r *QuestionRepository) UpdateQuestion(question *models.Question) error {
	return r.db.Model(question).Updates(question).Error
}

// UpdateQuestionPartial อัปเดตข้อมูลคำถามบางส่วน
func (r *QuestionRepository) UpdateQuestionPartial(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Question{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteQuestion ลบคำถาม
func (r *QuestionRepository) DeleteQuestion(id uint) error {
	return r.db.Delete(&models.Question{}, id).Error
}

// GetQuizIDByQuestionID ดึง QuizID จาก QuestionID
func (r *QuestionRepository) GetQuizIDByQuestionID(questionID uint) (uint, error) {
	var question models.Question
	err := r.db.Select("quiz_id").First(&question, questionID).Error
	if err != nil {
		return 0, err
	}
	return question.QuizID, nil
}