package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	dto "github.com/patiphanak/league-of-quiz/dto"
	models "github.com/patiphanak/league-of-quiz/model"
	"github.com/patiphanak/league-of-quiz/repositories"
)

// QuestionService สำหรับการจัดการ question
type QuestionService struct {
	questionRepo *repositories.QuestionRepository
	quizRepo     *repositories.QuizRepository
	choiceRepo   *repositories.ChoiceRepository
	fileService  *FileService
}

// NewQuestionService สร้าง instance ใหม่ของ QuestionService
func NewQuestionService(
	questionRepo *repositories.QuestionRepository,
	quizRepo *repositories.QuizRepository,
	fileService *FileService,
	choiceRepo *repositories.ChoiceRepository,
) *QuestionService {
	return &QuestionService{
		questionRepo: questionRepo,
		quizRepo:     quizRepo,
		fileService:  fileService,
		choiceRepo:  choiceRepo,
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

// UpdateQuestionWithChoices อัปเดตคำถามพร้อมตัวเลือกทั้งหมดในครั้งเดียว
// UpdateQuestionWithChoices อัปเดตคำถามพร้อมตัวเลือกทั้งหมดในครั้งเดียว
func (s *QuestionService) UpdateQuestionWithChoices(
	questionID uint,
	text string, 
	choices []dto.ChoiceFormData, 
	questionImage *multipart.FileHeader,
	choiceImages map[int]*multipart.FileHeader, 
	userID uint,
) error {
	// ตรวจสอบว่าคำถามมีอยู่จริง
	existingQuestion, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil {
		return err
	}

	// ตรวจสอบสิทธิ์ - ดึง QuizID และตรวจสอบความเป็นเจ้าของ
	quizID := existingQuestion.QuizID
	isOwner, err := s.quizRepo.CheckQuizOwnership(quizID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("unauthorized: you are not the owner of this quiz")
	}

	// 1. อัปเดตข้อมูลคำถาม
	questionToUpdate := &models.Question{
		ID:     questionID,
		QuizID: quizID, // ไม่อนุญาตให้เปลี่ยน QuizID
		Text:   text,
	}

	// จัดการไฟล์รูปภาพคำถาม
	if questionImage != nil {
		imageURL, err := s.fileService.UpdateFile(questionImage, existingQuestion.ImageURL, string(QuestionType))
		if err != nil {
			return err
		}
		questionToUpdate.ImageURL = imageURL
	} else {
		questionToUpdate.ImageURL = existingQuestion.ImageURL
	}

	// อัปเดตข้อมูลคำถาม
	if err := s.questionRepo.UpdateQuestion(questionToUpdate); err != nil {
		return err
	}

	// 2. ดึงตัวเลือกที่มีอยู่แล้ว
	existingChoices, err := s.choiceRepo.GetChoicesByQuestionID(questionID)
	if err != nil {
		return err
	}

	// สร้าง map เพื่อติดตามตัวเลือกที่อัปเดตแล้ว
	processedChoiceIds := make(map[uint]bool)

	// 3. จัดการตัวเลือก
	for i, choiceData := range choices {
		// ดึงไฟล์รูปภาพสำหรับตัวเลือกนี้ (ถ้ามี)
		choiceImage := choiceImages[i]
		
		if choiceData.ID == "" || choiceData.ID == "0" {
			// 3.1 สร้างตัวเลือกใหม่
			newChoice := &models.Choice{
				QuestionID: questionID,
				Text:       choiceData.Text,
				IsCorrect:  choiceData.IsCorrect,
			}
			
			// จัดการรูปภาพ
			if choiceImage != nil {
				imageURL, err := s.fileService.UploadFile(choiceImage, string(ChoiceType))
				if err != nil {
					return err
				}
				newChoice.ImageURL = imageURL
			}
			
			// สร้างตัวเลือกใหม่
			if err := s.choiceRepo.CreateChoice(newChoice); err != nil {
				return err
			}
		} else {
			// 3.2 อัปเดตตัวเลือกที่มีอยู่
			choiceID, err := strconv.ParseUint(choiceData.ID, 10, 32)
			if err != nil {
				return fmt.Errorf("invalid choice ID: %s", choiceData.ID)
			}
			
			// บันทึกว่าได้ประมวลผลตัวเลือกนี้แล้ว
			processedChoiceIds[uint(choiceID)] = true
			
			// ดึงข้อมูลตัวเลือกเดิม
			existingChoice, err := s.choiceRepo.GetChoiceByID(uint(choiceID))
			if err != nil {
				return fmt.Errorf("choice not found: %d", choiceID)
			}
			
			// สร้างข้อมูลสำหรับอัปเดต
			choiceToUpdate := &models.Choice{
				ID:         uint(choiceID),
				QuestionID: questionID, // คงค่าเดิม
				Text:       choiceData.Text,
				IsCorrect:  choiceData.IsCorrect,
			}
			
			// จัดการรูปภาพ
			if choiceImage != nil {
				imageURL, err := s.fileService.UpdateFile(choiceImage, existingChoice.ImageURL, string(ChoiceType))
				if err != nil {
					return err
				}
				choiceToUpdate.ImageURL = imageURL
			} else {
				choiceToUpdate.ImageURL = existingChoice.ImageURL
			}
			
			// อัปเดตตัวเลือก
			if err := s.choiceRepo.UpdateChoice(choiceToUpdate); err != nil {
				return err
			}
		}
	}

	// 4. ลบตัวเลือกที่ไม่ได้ส่งมา
	for _, existingChoice := range existingChoices {
		if !processedChoiceIds[existingChoice.ID] {
			// ลบรูปภาพ
			if existingChoice.ImageURL != "" {
				_ = s.fileService.DeleteFileByURL(existingChoice.ImageURL)
			}
			
			// ลบตัวเลือก
			if err := s.choiceRepo.DeleteChoice(existingChoice.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}