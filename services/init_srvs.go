package services

import (
	"log"
	"os"
	"path/filepath"

	"github.com/patiphanak/league-of-quiz/repositories"
)

// FileServiceOptions configurable options for FileService
type FileServiceOptions struct {
	BaseDir      string
	AllowedTypes []string
	MaxFileSize  int64
}

// DefaultFileOptions returns default options for FileService
func DefaultFileOptions(baseDir string) FileServiceOptions {
	return FileServiceOptions{
		BaseDir: baseDir,
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		MaxFileSize: 5 * 1024 * 1024, // 5MB
	}
}

// Services holds all service instances
type Services struct {
	Quiz *QuizService
	File *FileService
	// Other services here
}

// InitServices initializes all services with proper error handling
func InitServices(repos *repositories.Repositories, storagePath string) (*Services, error) {
	log.Println("Starting service initialization")

	// Create storage directory structure if it doesn't exist
	if err := ensureStorageDirectories(storagePath); err != nil {
		return nil, err
	}

	// Initialize file service with default options
	fileOptions := DefaultFileOptions(storagePath)
	fileService := NewFileService(fileOptions.BaseDir)
	fileService.allowedTypes = fileOptions.AllowedTypes
	fileService.maxFileSize = fileOptions.MaxFileSize

	log.Println("FileService initialized with storage at:", storagePath)

	// Initialize quiz service with its dependencies
	quizService := NewQuizService(repos.Quiz, fileService)
	log.Println("QuizService initialized successfully")

	// Create the services container
	services := &Services{
		Quiz: quizService,
		File: fileService,
		// Initialize other services here
	}

	log.Println("All services initialized successfully")
	return services, nil
}

// ensureStorageDirectories creates necessary storage directories
func ensureStorageDirectories(baseDir string) error {
	// Create main storage directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Printf("Failed to create base storage directory: %v", err)
		return err
	}

	// Create subdirectories for different file types
	for _, fileType := range []string{"quiz", "question", "choice"} {
		dirPath := filepath.Join(baseDir, fileType)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Printf("Failed to create %s directory: %v", fileType, err)
			return err
		}
	}

	return nil
}

// ShutdownServices performs cleanup for services
func (s *Services) ShutdownServices() {
	log.Println("Shutting down services...")
	// Perform any necessary cleanup here
	// For example, close connections, flush buffers, etc.
	log.Println("Services shutdown complete")
}
