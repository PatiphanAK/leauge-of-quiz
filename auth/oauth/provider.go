// auth/oauth/provider.go
package oauth

import (
	"context"
)

// Token เก็บข้อมูล OAuth token
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
}

// UserInfo เก็บข้อมูลผู้ใช้ที่ได้จาก OAuth provider
type UserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Provider เป็น interface สำหรับ OAuth providers
type Provider interface {
	// GetAuthURL สร้าง URL สำหรับการ redirect ไปยัง OAuth provider
	GetAuthURL(state string) string

	// Exchange แลกเปลี่ยน authorization code เพื่อรับ token
	Exchange(ctx context.Context, code string) (*Token, error)

	// GetUserInfo ดึงข้อมูลผู้ใช้จาก OAuth provider
	GetUserInfo(ctx context.Context, token *Token) (*UserInfo, error)
}
