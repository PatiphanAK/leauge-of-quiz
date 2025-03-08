package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileType ประเภทของไฟล์ที่จะอัปโหลด
type FileType string

const (
	QuizType     FileType = "quiz"
	QuestionType FileType = "question"
	ChoiceType   FileType = "choice"
)

// FileService สำหรับการจัดการไฟล์
type FileService struct {
	baseDir     string
	allowedTypes []string
	maxFileSize int64
}

// NewFileService สร้าง instance ใหม่ของ FileService
func NewFileService(baseDir string) *FileService {
	// สร้างโฟลเดอร์หลัก
	os.MkdirAll(baseDir, 0755)

	// สร้างโฟลเดอร์ย่อย
	for _, fileType := range []string{"quiz", "question", "choice"} {
		os.MkdirAll(filepath.Join(baseDir, fileType), 0755)
	}

	return &FileService{
		baseDir: baseDir,
		allowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		maxFileSize: 5 * 1024 * 1024, // 5MB
	}
}

// isAllowedFileType ตรวจสอบประเภทไฟล์
func (s *FileService) isAllowedFileType(contentType string) bool {
	for _, allowed := range s.allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// UploadFile อัปโหลดไฟล์และคืนค่า URL
func (s *FileService) UploadFile(file *multipart.FileHeader, fileType string) (string, error) {
	// ตรวจสอบขนาดไฟล์
	if file.Size > s.maxFileSize {
		return "", errors.New("file size exceeds the maximum limit of 5MB")
	}

	// ตรวจสอบประเภทไฟล์
	contentType := file.Header.Get("Content-Type")
	if !s.isAllowedFileType(contentType) {
		return "", errors.New("file type not allowed")
	}

	// สร้างชื่อไฟล์ไม่ซ้ำ
	filename := uuid.New().String() + filepath.Ext(file.Filename)

	// สร้างเส้นทางโฟลเดอร์
	uploadDir := filepath.Join(s.baseDir, fileType)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// บันทึกไฟล์
	dst := filepath.Join(uploadDir, filename)
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// อ่านข้อมูลจาก source และเขียนไปยัง destination
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return "", err
		}
		if n == 0 {
			break
		}

		if _, err := out.Write(buf[:n]); err != nil {
			return "", err
		}
	}

	// สร้าง URL สำหรับเข้าถึงไฟล์
	fileURL := fmt.Sprintf("/upload/%s/%s", fileType, filename)

	return fileURL, nil
}

// DeleteFile ลบไฟล์
func (s *FileService) DeleteFile(filename string, fileType string) error {
	if filename == "" {
		return nil
	}

	// สร้างเส้นทางไฟล์
	filePath := filepath.Join(s.baseDir, fileType, filename)

	// ตรวจสอบว่าไฟล์มีอยู่หรือไม่
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // ไม่มีไฟล์ ไม่ต้องทำอะไรต่อ
	}

	// ลบไฟล์
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFilePath คืนค่าเส้นทางไฟล์จาก URL
func (s *FileService) GetFilePath(fileURL string) (string, string, error) {
	// ตรวจสอบว่า URL มีรูปแบบที่ถูกต้องหรือไม่
	if !strings.HasPrefix(fileURL, "/upload/") {
		return "", "", errors.New("invalid file URL format")
	}

	// แยกส่วนของ URL
	parts := strings.Split(strings.TrimPrefix(fileURL, "/upload/"), "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid file URL format")
	}

	fileType := parts[0]
	filename := parts[1]

	// สร้างเส้นทางไฟล์
	filePath := filepath.Join(s.baseDir, fileType, filename)

	return filePath, fileType, nil
}

// UpdateFile อัปเดตไฟล์โดยลบไฟล์เก่าและอัปโหลดไฟล์ใหม่
func (s *FileService) UpdateFile(file *multipart.FileHeader, oldFileURL string, fileType string) (string, error) {
	// ถ้ามีไฟล์เก่า ให้ลบออก
	if oldFileURL != "" {
		// แยกชื่อไฟล์จาก URL
		oldFilename := strings.TrimPrefix(oldFileURL, fmt.Sprintf("/upload/%s/", fileType))
		if oldFilename != "" {
			if err := s.DeleteFile(oldFilename, fileType); err != nil {
				// ไม่ต้องคืนค่าข้อผิดพลาด เพราะเราต้องการอัปโหลดไฟล์ใหม่ต่อไป
				fmt.Printf("Warning: Failed to delete old file: %v\n", err)
			}
		}
	}

	// อัปโหลดไฟล์ใหม่
	return s.UploadFile(file, fileType)
}