package services

import (
	"errors"
	"mime/multipart"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// QuestionService สำหรับการจัดการ question
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
func (s *QuestionService) CreateQuestion(question *models.Question, imageFile *multipart.FileHeader, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(question.QuizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		imageURL, err := s.fileService.UploadFile(imageFile, string(QuestionType))
		if err != nil {
			return err
		}
		question.ImageURL = imageURL
	}

	// บันทึกข้อมูลคำถาม
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
func (s *QuestionService) UpdateQuestion(question *models.Question, imageFile *multipart.FileHeader, currentUserID uint) error {
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

	// ดึงข้อมูลคำถามเดิม
	existingQuestion, err := s.questionRepo.GetQuestionByID(question.ID)
	if err != nil {
		return err
	}
	
	// ป้องกันการเปลี่ยน QuizID
	question.QuizID = existingQuestion.QuizID

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		imageURL, err := s.fileService.UpdateFile(imageFile, existingQuestion.ImageURL, string(QuestionType))
		if err != nil {
			return err
		}
		question.ImageURL = imageURL
	} else {
		// ถ้าไม่มีการอัปโหลดรูปภาพใหม่ ให้ใช้รูปภาพเดิม
		question.ImageURL = existingQuestion.ImageURL
	}

	// อัปเดตข้อมูลคำถาม
	return s.questionRepo.UpdateQuestion(question)
}

// PatchQuestion อัปเดตข้อมูลคำถามบางส่วน
func (s *QuestionService) PatchQuestion(questionID uint, updates map[string]interface{}, imageFile *multipart.FileHeader, currentUserID uint) error {
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
	delete(updates, "quiz_id")

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		// ดึงข้อมูลคำถามเดิม
		existingQuestion, err := s.questionRepo.GetQuestionByID(questionID)
		if err != nil {
			return err
		}

		imageURL, err := s.fileService.UpdateFile(imageFile, existingQuestion.ImageURL, string(QuestionType))
		if err != nil {
			return err
		}
		updates["image_url"] = imageURL
	}

	// อัปเดตข้อมูลคำถาม
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
		_ = s.fileService.DeleteFileByURL(question.ImageURL)
	}

	// ลบข้อมูลคำถาม
	return s.questionRepo.DeleteQuestion(questionID)
}