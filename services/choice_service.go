package services

import (
	"errors"
	"mime/multipart"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// ChoiceService สำหรับการจัดการตัวเลือก
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
func (s *ChoiceService) CreateChoice(choice *models.Choice, imageFile *multipart.FileHeader, currentUserID uint) error {
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

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		imageURL, err := s.fileService.UploadFile(imageFile, string(ChoiceType))
		if err != nil {
			return err
		}
		choice.ImageURL = imageURL
	}

	// บันทึกข้อมูลตัวเลือก
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
func (s *ChoiceService) UpdateChoice(choice *models.Choice, imageFile *multipart.FileHeader, currentUserID uint) error {
	// ดึงข้อมูลตัวเลือกเดิม
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

	// ป้องกันการเปลี่ยน QuestionID
	choice.QuestionID = existingChoice.QuestionID

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		imageURL, err := s.fileService.UpdateFile(imageFile, existingChoice.ImageURL, string(ChoiceType))
		if err != nil {
			return err
		}
		choice.ImageURL = imageURL
	} else {
		// ถ้าไม่มีการอัปโหลดรูปภาพใหม่ ให้ใช้รูปภาพเดิม
		choice.ImageURL = existingChoice.ImageURL
	}

	// อัปเดตข้อมูลตัวเลือก
	return s.choiceRepo.UpdateChoice(choice)
}

// DeleteChoice ลบตัวเลือก
func (s *ChoiceService) DeleteChoice(choiceID uint, currentUserID uint) error {
	// ดึงข้อมูลตัวเลือก
	choice, err := s.choiceRepo.GetChoiceByID(choiceID)
	if err != nil {
		return err
	}

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

	// ลบรูปภาพของตัวเลือกถ้ามี
	if choice.ImageURL != "" {
		_ = s.fileService.DeleteFileByURL(choice.ImageURL)
	}

	// ลบข้อมูลตัวเลือก
	return s.choiceRepo.DeleteChoice(choiceID)
}