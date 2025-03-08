package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/patiphanak/league-of-quiz/config"
	models "github.com/patiphanak/league-of-quiz/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims โครงสร้าง claims สำหรับ JWT
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// Service สำหรับจัดการ JWT
type JWTService struct {
	config *config.Config
}

// NewJWTService สร้าง instance ใหม่ของ JWTService
func NewJWTService(config *config.Config) *JWTService {
	return &JWTService{
		config: config,
	}
}

// GenerateToken สร้าง JWT token สำหรับผู้ใช้
func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	// สร้าง claims
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // หมดอายุใน 24 ชั่วโมง
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// สร้าง token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// ลงนาม token ด้วย secret key
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken ตรวจสอบความถูกต้องของ token
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// ตรวจสอบว่า signing method เป็น HMAC หรือไม่
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// ดึง claims จาก token
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
