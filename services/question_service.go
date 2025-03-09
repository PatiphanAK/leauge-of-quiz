package services

import (
	"errors"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

type QuestionService struct {
	questionRepo *repositories.QuestionRepository
	quizRepo     *repositories.QuizRepository
	fileService  *FileService
}

// NewQuestionService สร้าง instance ใหม่ของ QuestionService
func NewQuestionService(
	questionRepo *repositories.QuestionRepository,
	quizRepo *repositories.QuizRepository,
	fileService *FileService,
) *QuestionService {
	return &QuestionService{
		questionRepo: questionRepo,
		quizRepo:     quizRepo,
		fileService:  fileService,
	}
}

// CreateQuestion สร้างคำถามใหม่
func (s *QuestionService) CreateQuestion(question *models.Question, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(question.QuizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	return s.questionRepo.CreateQuestion(question)
}

// GetQuestionByID ดึงข้อมูลคำถามจาก ID
func (s *QuestionService) GetQuestionByID(id uint) (*models.Question, error) {
	return s.questionRepo.GetQuestionByID(id)
}

// GetQuestionsByQuizID ดึงคำถามทั้งหมดของ quiz
func (s *QuestionService) GetQuestionsByQuizID(quizID uint) ([]models.Question, error) {
	return s.questionRepo.GetQuestionsByQuizID(quizID)
}

// UpdateQuestion อัปเดตข้อมูลคำถาม
func (s *QuestionService) UpdateQuestion(question *models.Question, currentUserID uint) error {
	// ดึง QuizID จาก QuestionID
	quizID, err := s.questionRepo.GetQuizIDByQuestionID(question.ID)
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

	// ป้องกันการเปลี่ยน QuizID
	existingQuestion, err := s.questionRepo.GetQuestionByID(question.ID)
	if err != nil {
		return err
	}
	question.QuizID = existingQuestion.QuizID

	return s.questionRepo.UpdateQuestion(question)
}

// PatchQuestion อัปเดตข้อมูลคำถามบางส่วน
func (s *QuestionService) PatchQuestion(questionID uint, updates map[string]interface{}, currentUserID uint) error {
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

	// ป้องกันการเปลี่ยน QuizID
	if _, exists := updates["quiz_id"]; exists {
		delete(updates, "quiz_id")
	}

	return s.questionRepo.UpdateQuestionPartial(questionID, updates)
}

// DeleteQuestion ลบคำถาม
func (s *QuestionService) DeleteQuestion(questionID uint, currentUserID uint) error {
	// ดึงข้อมูลคำถาม
	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(question.QuizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ลบรูปภาพของคำถามถ้ามี
	if question.ImageURL != "" {
		filePath, fileType, err := s.fileService.GetFilePath(question.ImageURL)
		if err == nil {
			_ = s.fileService.DeleteFile(filePath, fileType)
		}
	}

	// ลบรูปภาพของตัวเลือก
	for _, choice := range question.Choices {
		if choice.ImageURL != "" {
			filePath, fileType, err := s.fileService.GetFilePath(choice.ImageURL)
			if err == nil {
				_ = s.fileService.DeleteFile(filePath, fileType)
			}
		}
	}

	return s.questionRepo.DeleteQuestion(questionID)
}
