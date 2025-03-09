package services

import (
	"errors"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

type QuizService struct {
	quizRepo    *repositories.QuizRepository
	fileService *FileService
}

// NewQuizService สร้าง instance ใหม่ของ QuizService
func NewQuizService(quizRepo *repositories.QuizRepository, fileService *FileService) *QuizService {
	return &QuizService{
		quizRepo:    quizRepo,
		fileService: fileService,
	}
}

// CreateQuiz สร้าง quiz ใหม่
func (s *QuizService) CreateQuiz(quiz *models.Quiz) error {
	return s.quizRepo.CreateQuiz(quiz)
}

// GetQuizByID ดึงข้อมูล quiz จาก ID
func (s *QuizService) GetQuizByID(id uint) (*models.Quiz, error) {
	return s.quizRepo.GetQuizByID(id)
}

// GetAllQuizzes ดึงข้อมูล quizzes ทั้งหมด
func (s *QuizService) GetAllQuizzes(page, limit int) ([]models.Quiz, int64, error) {
	return s.quizRepo.GetAllQuizzes(page, limit)
}

// GetPublishedQuizzes ดึงข้อมูล quizzes ที่เผยแพร่แล้ว
func (s *QuizService) GetPublishedQuizzes(page, limit int) ([]models.Quiz, int64, error) {
	return s.quizRepo.GetPublishedQuizzes(page, limit)
}

// GetQuizzesByCreator ดึงข้อมูล quizzes ของผู้สร้าง
func (s *QuizService) GetQuizzesByCreator(creatorID uint, page, limit int) ([]models.Quiz, int64, error) {
	return s.quizRepo.GetQuizzesByCreator(creatorID, page, limit)
}

// UpdateQuiz อัปเดตข้อมูล quiz
func (s *QuizService) UpdateQuiz(quiz *models.Quiz, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quiz.ID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ตรวจสอบให้แน่ใจว่า CreatorID ไม่ถูกเปลี่ยน
	existingQuiz, err := s.quizRepo.GetQuizByID(quiz.ID)
	if err != nil {
		return err
	}
	quiz.CreatorID = existingQuiz.CreatorID

	return s.quizRepo.UpdateQuiz(quiz)
}

// PatchQuiz อัปเดตข้อมูล quiz บางส่วน
func (s *QuizService) PatchQuiz(quizID uint, updates map[string]interface{}, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ป้องกันการเปลี่ยน CreatorID
	if _, exists := updates["creator_id"]; exists {
		delete(updates, "creator_id")
	}

	// สร้าง quiz object เพื่อที่จะ update
	quiz := &models.Quiz{
		ID: quizID,
	}

	return s.quizRepo.UpdateQuiz(quiz)
}

// DeleteQuiz ลบ quiz
func (s *QuizService) DeleteQuiz(quizID uint, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ดึงข้อมูล quiz เพื่อลบรูปภาพที่เกี่ยวข้อง
	quiz, err := s.quizRepo.GetQuizByID(quizID)
	if err != nil {
		return err
	}

	// ลบรูปภาพของ quiz ถ้ามี
	if quiz.ImageURL != "" {
		filePath, fileType, err := s.fileService.GetFilePath(quiz.ImageURL)
		if err == nil {
			// ถ้าลบรูปภาพไม่สำเร็จ ให้ทำการลบ quiz ต่อไป
			_ = s.fileService.DeleteFile(filePath, fileType)
		}
	}

	// ลบรูปภาพของคำถามและตัวเลือก
	for _, question := range quiz.Questions {
		if question.ImageURL != "" {
			filePath, fileType, err := s.fileService.GetFilePath(question.ImageURL)
			if err == nil {
				_ = s.fileService.DeleteFile(filePath, fileType)
			}
		}

		for _, choice := range question.Choices {
			if choice.ImageURL != "" {
				filePath, fileType, err := s.fileService.GetFilePath(choice.ImageURL)
				if err == nil {
					_ = s.fileService.DeleteFile(filePath, fileType)
				}
			}
		}
	}

	return s.quizRepo.DeleteQuiz(quizID)
}

// UpdateQuizCategories อัปเดตหมวดหมู่ของ quiz
func (s *QuizService) UpdateQuizCategories(quizID uint, categoryIDs []uint, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	return s.quizRepo.UpdateQuizCategories(quizID, categoryIDs)
}