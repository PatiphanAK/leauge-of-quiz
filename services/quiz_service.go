package services

import (
	"errors"
	"log"

	models "github.com/patiphanak/league-of-quiz/model"
	repositories "github.com/patiphanak/league-of-quiz/repositories"
)

type QuizService struct {
	quizRepo     *repositories.QuizRepository
	questionRepo *repositories.QuestionRepository
	choiceRepo   *repositories.ChoiceRepository
	fileService  *FileService
}

func NewQuizService(
	quizRepo *repositories.QuizRepository,
	questionRepo *repositories.QuestionRepository,
	choiceRepo *repositories.ChoiceRepository,
	fileService *FileService,
) *QuizService {
	log.Println("NewQuizService")
	return &QuizService{
		quizRepo:     quizRepo,
		questionRepo: questionRepo,
		choiceRepo:   choiceRepo,
		fileService:  fileService,
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
	log.Println("get  service of all quizzes")
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
	// Ownership check
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, currentUserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// Prevent changing CreatorID
	delete(updates, "creator_id")

	// Now call the repository with map updates
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

// QuestionData สำหรับข้อมูลคำถาม
type QuestionData struct {
	Text     string       `json:"text"`
	ImageURL string       `json:"imageURL"`
	Choices  []ChoiceData `json:"choices"`
}

// ChoiceData สำหรับข้อมูลตัวเลือก
type ChoiceData struct {
	Text      string `json:"text"`
	ImageURL  string `json:"imageURL"`
	IsCorrect bool   `json:"isCorrect"`
}

// CreateQuizWithQuestionsAndChoices สร้าง quiz พร้อมคำถามและตัวเลือก
func (s *QuizService) CreateQuizWithQuestionsAndChoices(quiz *models.Quiz, questions []QuestionData, categories []uint, userID uint) error {
	// เริ่ม transaction
	tx := s.quizRepo.GetDB().Begin()

	// สร้าง quiz
	if err := tx.Create(quiz).Error; err != nil {
		tx.Rollback()
		return err
	}

	// สร้างความสัมพันธ์กับหมวดหมู่ (ถ้ามี)
	if len(categories) > 0 {
		for _, catID := range categories {
			if err := tx.Exec("INSERT INTO quiz_categories (quiz_id, category_id) VALUES (?, ?)", quiz.ID, catID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// สร้างคำถามและตัวเลือก
	for _, qData := range questions {
		question := models.Question{
			QuizID:   quiz.ID,
			Text:     qData.Text,
			ImageURL: qData.ImageURL,
		}

		if err := tx.Create(&question).Error; err != nil {
			tx.Rollback()
			return err
		}

		for _, cData := range qData.Choices {
			choice := models.Choice{
				QuestionID: question.ID,
				Text:       cData.Text,
				ImageURL:   cData.ImageURL,
				IsCorrect:  cData.IsCorrect,
			}

			if err := tx.Create(&choice).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Commit transaction
	return tx.Commit().Error
}

func (s *QuizService) GetFilteredQuizzes(offset, limit int, isPublished string, search string, categories []uint) ([]models.Quiz, int64, error) {
	return s.quizRepo.GetFilteredQuizzes(offset, limit, isPublished, search, categories)
}

func (s *QuizService) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	err := s.quizRepo.GetDB().Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}
