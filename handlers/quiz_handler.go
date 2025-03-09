package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/patiphanak/league-of-quiz/services"
)

// UploadHandler สำหรับจัดการการอัปโหลดไฟล์
type UploadHandler struct {
	fileService *services.FileService
}

// NewUploadHandler สร้าง instance ใหม่ของ UploadHandler
func NewUploadHandler(fileService *services.FileService) *UploadHandler {
	return &UploadHandler{
		fileService: fileService,
	}
}

// validateFileType ตรวจสอบประเภทของไฟล์
func validateFileType(fileType string) (string, error) {
	switch fileType {
	case "quiz", "question", "choice":
		return fileType, nil
	default:
		return "", fmt.Errorf("invalid file type: %s", fileType)
	}
}

// UploadFile อัปโหลดไฟล์
func (h *UploadHandler) UploadFile(c *fiber.Ctx) error {
	// รับประเภทของไฟล์จาก path parameter
	fileType := c.Params("type")
	
	// ตรวจสอบประเภทของไฟล์
	validFileType, err := validateFileType(fileType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// รับไฟล์
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided or invalid file",
		})
	}

	// ถ้ามี old_file_url ให้อัปเดตแทนการสร้างใหม่
	oldFileURL := c.FormValue("old_file_url", "")

	var fileURL string
	if oldFileURL != "" {
		fileURL, err = h.fileService.UpdateFile(file, oldFileURL, validFileType)
	} else {
		fileURL, err = h.fileService.UploadFile(file, validFileType)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"url": fileURL,
	})
}

// DeleteFile ลบไฟล์
func (h *UploadHandler) DeleteFile(c *fiber.Ctx) error {
	// รับประเภทของไฟล์จาก path parameter
	fileType := c.Params("type")
	
	// ตรวจสอบประเภทของไฟล์
	validFileType, err := validateFileType(fileType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// รับชื่อไฟล์จาก parameter
	filename := c.Params("filename")
	if filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No filename provided",
		})
	}

	// ลบไฟล์
	if err := h.fileService.DeleteFile(filename, validFileType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}