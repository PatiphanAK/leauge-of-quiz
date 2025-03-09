package repositories

import (
	models "github.com/patiphanak/league-of-quiz/model"
	"gorm.io/gorm"
)

type ChoiceRepository struct {
	db *gorm.DB
}

// NewChoiceRepository สร้าง instance ใหม่ของ ChoiceRepository
func NewChoiceRepository(db *gorm.DB) *ChoiceRepository {
	return &ChoiceRepository{db: db}
}

// CreateChoice สร้างตัวเลือกใหม่
func (r *ChoiceRepository) CreateChoice(choice *models.Choice) error {
	return r.db.Create(choice).Error
}

// GetChoiceByID ดึงข้อมูลตัวเลือกจาก ID
func (r *ChoiceRepository) GetChoiceByID(id uint) (*models.Choice, error) {
	var choice models.Choice
	err := r.db.First(&choice, id).Error
	if err != nil {
		return nil, err
	}
	return &choice, nil
}

// GetChoicesByQuestionID ดึงตัวเลือกทั้งหมดของคำถาม
func (r *ChoiceRepository) GetChoicesByQuestionID(questionID uint) ([]models.Choice, error) {
	var choices []models.Choice
	err := r.db.Where("question_id = ?", questionID).Find(&choices).Error
	if err != nil {
		return nil, err
	}
	return choices, nil
}

// UpdateChoice อัปเดตข้อมูลตัวเลือก
func (r *ChoiceRepository) UpdateChoice(choice *models.Choice) error {
	return r.db.Model(choice).Updates(choice).Error
}

// UpdateChoicePartial อัปเดตข้อมูลตัวเลือกบางส่วน
func (r *ChoiceRepository) UpdateChoicePartial(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Choice{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteChoice ลบตัวเลือก
func (r *ChoiceRepository) DeleteChoice(id uint) error {
	return r.db.Delete(&models.Choice{}, id).Error
}

// GetQuestionIDByChoiceID ดึง QuestionID จาก ChoiceID
func (r *ChoiceRepository) GetQuestionIDByChoiceID(choiceID uint) (uint, error) {
	var choice models.Choice
	err := r.db.Select("question_id").First(&choice, choiceID).Error
	if err != nil {
		return 0, err
	}
	return choice.QuestionID, nil
}