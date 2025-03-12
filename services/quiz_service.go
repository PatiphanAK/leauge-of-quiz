package services

import (
	"errors"
	"mime/multipart"

	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// QuizService สำหรับการจัดการ quiz
type QuizService struct {
	quizRepo    *repositories.QuizRepository
	fileService *FileService
}

// NewQuizService สร้าง instance ใหม่ของ QuizService
func NewQuizService(
	quizRepo *repositories.QuizRepository,
	fileService *FileService,
) *QuizService {
	return &QuizService{
		quizRepo:    quizRepo,
		fileService: fileService,
	}
}

// CreateQuiz สร้าง quiz ใหม่
func (s *QuizService) CreateQuiz(quiz *models.Quiz, imageFile *multipart.FileHeader) error {
	// จัดการไฟล์รูปภาพก่อน
	if imageFile != nil {
		imageURL, err := s.fileService.UploadFile(imageFile, string(QuizType))
		if err != nil {
			return err
		}
		quiz.ImageURL = imageURL
	}

	// บันทึกข้อมูล quiz
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
func (s *QuizService) UpdateQuiz(quiz *models.Quiz, imageFile *multipart.FileHeader, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quiz.ID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ดึงข้อมูล quiz เดิม
	existingQuiz, err := s.quizRepo.GetQuizByID(quiz.ID)
	if err != nil {
		return err
	}
	
	// ป้องกันการเปลี่ยน CreatorID
	quiz.CreatorID = existingQuiz.CreatorID

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		imageURL, err := s.fileService.UpdateFile(imageFile, existingQuiz.ImageURL, string(QuizType))
		if err != nil {
			return err
		}
		quiz.ImageURL = imageURL
	} else {
		// ถ้าไม่มีการอัปโหลดรูปภาพใหม่ ให้ใช้รูปภาพเดิม
		quiz.ImageURL = existingQuiz.ImageURL
	}

	// อัปเดตข้อมูล quiz
	return s.quizRepo.UpdateQuiz(quiz)
}

// PatchQuiz อัปเดตข้อมูล quiz บางส่วน
func (s *QuizService) PatchQuiz(quizID uint, updates map[string]interface{}, imageFile *multipart.FileHeader, currentUserID uint) error {
	// ตรวจสอบว่าผู้ใช้เป็นเจ้าของ quiz หรือไม่
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// ป้องกันการเปลี่ยน CreatorID
	delete(updates, "creator_id")

	// จัดการไฟล์รูปภาพ
	if imageFile != nil {
		// ดึงข้อมูล quiz เดิม
		existingQuiz, err := s.quizRepo.GetQuizByID(quizID)
		if err != nil {
			return err
		}

		imageURL, err := s.fileService.UpdateFile(imageFile, existingQuiz.ImageURL, string(QuizType))
		if err != nil {
			return err
		}
		updates["image_url"] = imageURL
	}

	// อัปเดตข้อมูล quiz
	return s.quizRepo.UpdateQuizWithMap(quizID, updates)
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
		_ = s.fileService.DeleteFileByURL(quiz.ImageURL)
	}

	// ลบ quiz จากฐานข้อมูล
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

// GetFilteredQuizzes ดึง quiz ตามเงื่อนไขการกรอง
func (s *QuizService) GetFilteredQuizzes(offset, limit int, isPublished string, search string, categories []uint) ([]models.Quiz, int64, error) {
	return s.quizRepo.GetFilteredQuizzes(offset, limit, isPublished, search, categories)
}

// GetAllCategories ดึงหมวดหมู่ทั้งหมด
func (s *QuizService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := s.quizRepo.GetDB().Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}