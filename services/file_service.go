package services

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
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
	baseDir      string
	allowedTypes []string
	maxFileSize  int64
	serverURL    string // เพิ่ม URL ของเซิร์ฟเวอร์
}

// NewFileService สร้าง instance ใหม่ของ FileService
func NewFileService(baseDir string) *FileService {
	return &FileService{
		baseDir: baseDir,
		allowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		maxFileSize: 5 * 1024 * 1024,  // 5MB
		serverURL:   "localhost:3000", // เก็บ URL ของเซิร์ฟเวอร์
	}
}

// IsAllowedFileType ตรวจสอบประเภทไฟล์
func (s *FileService) IsAllowedFileType(contentType string) bool {
	for _, allowed := range s.allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// UploadFile อัปโหลดไฟล์และคืนค่า URL
func (s *FileService) UploadFile(file *multipart.FileHeader, fileType string) (string, error) {
	if file == nil {
		return "", nil
	}

	// ตรวจสอบว่า fileType ถูกต้อง
	if fileType != string(QuizType) && fileType != string(QuestionType) && fileType != string(ChoiceType) {
		return "", errors.New("invalid file type category")
	}

	// ตรวจสอบขนาดไฟล์
	if file.Size > s.maxFileSize {
		return "", errors.New("file size exceeds the maximum limit of 5MB")
	}

	// ตรวจสอบประเภทไฟล์
	contentType := file.Header.Get("Content-Type")
	if !s.IsAllowedFileType(contentType) {
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

	// สร้าง URL ในรูปแบบที่ถูกต้อง
	fileURL := fmt.Sprintf("%s/storage/%s/%s", s.serverURL, fileType, filename)

	return fileURL, nil
}

// DeleteFile ลบไฟล์จากชื่อไฟล์และประเภท
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

// DeleteFileByURL ลบไฟล์จาก URL
func (s *FileService) DeleteFileByURL(fileURL string) error {
	if fileURL == "" {
		return nil
	}

	// แยกส่วนของ URL
	filename, fileType, err := s.ExtractInfoFromURL(fileURL)
	if err != nil {
		return err
	}
	return s.DeleteFile(filename, fileType)
}

// GetFilePath คืนค่าเส้นทางไฟล์จาก URL
func (s *FileService) GetFilePath(fileURL string) (string, string, error) {
	// แปลงรูปแบบ URL ให้ถูกต้อง
	fileURL = strings.TrimPrefix(fileURL, s.serverURL)

	// ตรวจสอบว่า URL มีรูปแบบที่ถูกต้องหรือไม่
	if !strings.HasPrefix(fileURL, "/storage/") {
		return "", "", errors.New("invalid file URL format")
	}

	// แยกส่วนของ URL
	parts := strings.Split(strings.TrimPrefix(fileURL, "/storage/"), "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid file URL format")
	}

	fileType := parts[0]
	filename := parts[1]

	// สร้างเส้นทางไฟล์
	filePath := filepath.Join(s.baseDir, fileType, filename)

	return filePath, fileType, nil
}

func (s *FileService) ExtractInfoFromURL(fileURL string) (string, string, error) {
	if fileURL == "" {
		return "", "", nil
	}

	// Regular expression to extract file type and filename from URL
	// Pattern matches: [anything]/storage/[filetype]/[filename]
	re := regexp.MustCompile(`/storage/([^/]+)/([^/]+)$`)
	matches := re.FindStringSubmatch(fileURL)

	if matches == nil || len(matches) < 3 {
		return "", "", fmt.Errorf("invalid file URL format: %s", fileURL)
	}

	fileType := matches[1] // e.g., "quiz"
	filename := matches[2] // e.g., "5f93c498-050d-4fb6-8a0f-416ec7d1876f.jpg"
	return filename, fileType, nil
}

// UpdateFile อัปเดตไฟล์โดยลบไฟล์เก่าและอัปโหลดไฟล์ใหม่
func (s *FileService) UpdateFile(file *multipart.FileHeader, oldFileURL string, fileType string) (string, error) {
	// ถ้าไม่มีไฟล์ใหม่และไม่มี URL เก่า
	if file == nil {
		return oldFileURL, nil
	}

	// ถ้ามีไฟล์เก่า ให้ลบออก
	if oldFileURL != "" {
		err := s.DeleteFileByURL(oldFileURL)
		if err != nil {
			log.Printf("WARNING: Failed to delete old file: %v", err)
			// Continue anyway since we still want to upload the new file
		} else {
			log.Printf("Successfully deleted old file: %s", oldFileURL)
		}
	}

	// อัปโหลดไฟล์ใหม่
	return s.UploadFile(file, fileType)
}
