package services

import (
	"errors"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

type ChoiceService struct {
	choiceRepo   *repositories.ChoiceRepository
	questionRepo *repositories.QuestionRepository
	quizRepo     *repositories.QuizRepository
	fileService  *FileService
}

// NewChoiceService สร้าง instance ใหม่ของ ChoiceService
func NewChoiceService(
	choiceRepo *repositories.ChoiceRepository,
	questionRepo *repositories.QuestionRepository,
	quizRepo *repositories.QuizRepository,
	fileService *FileService,
) *ChoiceService {
	return &ChoiceService{
		choiceRepo:   choiceRepo,
		questionRepo: questionRepo,
		quizRepo:     quizRepo,
		fileService:  fileService,
	}
}

// CreateChoice สร้างตัวเลือกใหม่
func (s *ChoiceService) CreateChoice(choice *models.Choice, currentUserID uint) error {
	// ดึง QuizID จาก QuestionID
	quizID, err := s.questionRepo.GetQuizIDByQuestionID(choice.QuestionID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	return s.choiceRepo.CreateChoice(choice)
}

// GetChoiceByID ดึงข้อมูลตัวเลือกจาก ID
func (s *ChoiceService) GetChoiceByID(id uint) (*models.Choice, error) {
	return s.choiceRepo.GetChoiceByID(id)
}

// GetChoicesByQuestionID ดึงตัวเลือกทั้งหมดของคำถาม
func (s *ChoiceService) GetChoicesByQuestionID(questionID uint) ([]models.Choice, error) {
	return s.choiceRepo.GetChoicesByQuestionID(questionID)
}

// UpdateChoice อัปเดตข้อมูลตัวเลือก
func (s *ChoiceService) UpdateChoice(choice *models.Choice, currentUserID uint) error {
	// ดึง QuestionID จาก ChoiceID
	existingChoice, err := s.choiceRepo.GetChoiceByID(choice.ID)
	if err != nil {
		return err
	}

	// ดึง QuizID จาก QuestionID
	quizID, err := s.questionRepo.GetQuizIDByQuestionID(existingChoice.QuestionID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}
	return s.choiceRepo.UpdateChoice(choice)
}

// UpdateChoicePartial อัปเดตข้อมูลตัวเลือกบางส่วน
func (s *ChoiceService) UpdateChoicePartial(id uint, updates map[string]interface{}, currentUserID uint) error {
	// ดึง QuestionID จาก ChoiceID
	questionID, err := s.choiceRepo.GetQuestionIDByChoiceID(id)
	if err != nil {
		return err
	}

	// ดึง QuizID จาก QuestionID
	quizID, err := s.questionRepo.GetQuizIDByQuestionID(questionID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}
	return s.choiceRepo.UpdateChoicePartial(id, updates)
}

// DeleteChoice ลบตัวเลือก
func (s *ChoiceService) DeleteChoice(id uint, currentUserID uint) error {
	// ดึง QuestionID จาก ChoiceID
	questionID, err := s.choiceRepo.GetQuestionIDByChoiceID(id)
	if err != nil {
		return err
	}

	// ดึง QuizID จาก QuestionID
	quizID, err := s.questionRepo.GetQuizIDByQuestionID(questionID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}
	return s.choiceRepo.DeleteChoice(id)
}