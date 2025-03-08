package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	JWTSecret          string
	JWTRefreshSecret   string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	return &Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTRefreshSecret:   os.Getenv("JWT_REFRESH_SECRET"),
	}, nil
}
