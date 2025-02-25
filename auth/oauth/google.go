package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/patiphanak/league-of-quiz/config"
)

const (
	googleAuthURL     = "https://accounts.google.com/o/oauth2/auth"
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleOAuth implementation ของ Provider interface สำหรับ Google
type GoogleOAuth struct {
	config *config.Config
}

// NewGoogleOAuth สร้าง instance ใหม่ของ GoogleOAuth
func NewGoogleOAuth(config *config.Config) *GoogleOAuth {
	return &GoogleOAuth{
		config: config,
	}
}

// GetAuthURL implements Provider
func (g *GoogleOAuth) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", g.config.GoogleClientID)
	params.Add("redirect_uri", g.config.GoogleRedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid profile email")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")
	authURL := fmt.Sprintf("%s?%s", googleAuthURL, params.Encode())
	fmt.Println(authURL)
	return authURL
}

// Exchange implements Provider
func (g *GoogleOAuth) Exchange(ctx context.Context, code string) (*Token, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", g.config.GoogleClientID)
	data.Set("client_secret", g.config.GoogleClientSecret)
	data.Set("redirect_uri", g.config.GoogleRedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", googleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchanging code for token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token response error: %s", body)
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	return &token, nil
}

// GetUserInfo implements Provider
func (g *GoogleOAuth) GetUserInfo(ctx context.Context, token *Token) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", googleUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating user info request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading user info response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info response error: %s", body)
	}

	var userInfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		EmailVerified bool   `json:"email_verified"`
	}

	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("parsing user info: %w", err)
	}

	return &UserInfo{
		ID:      userInfo.Sub,
		Email:   userInfo.Email,
		Name:    userInfo.Name,
		Picture: userInfo.Picture,
	}, nil
}
