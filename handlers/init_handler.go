package handlers

import (
	"github.com/patiphanak/league-of-quiz/auth/jwt"
	"github.com/patiphanak/league-of-quiz/auth/oauth"
	"github.com/patiphanak/league-of-quiz/services"
	"gorm.io/gorm"
)

// AllHandlers รวบรวม handler ทั้งหมด
type AllHandlers struct {
	Auth     *AuthHandler
	Quiz     *QuizHandler
	Upload   *UploadHandler
	Question *QuestionHandler
	Choice   *ChoiceHandler
}

// InitHandlers สร้าง instance ทั้งหมดของ handlers
func InitHandlers(
	services *services.Services,
	db *gorm.DB,
	jwtService *jwt.JWTService,
	googleOAuth *oauth.GoogleOAuth,
) *AllHandlers {
	return &AllHandlers{
		Auth:     NewAuthHandler(db, googleOAuth, jwtService),
		Quiz:     NewQuizHandler(services.Quiz, services.File),
		Upload:   NewUploadHandler(services.File),
		Question: NewQuestionHandler(services.Question, services.File),
		Choice:   NewChoiceHandler(services.Choice, services.File),
	}
}