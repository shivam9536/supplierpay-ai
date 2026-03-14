package config

import (
	"os"
	"strconv"
)

type Config struct {
	// App
	AppEnv      string
	AppPort     string
	FrontendURL string
	MockMode    bool

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// AWS General
	AWSRegion string

	// Bedrock
	BedrockModelID      string
	BedrockMaxTokens    int
	BedrockBearerToken  string

	// S3
	S3BucketName string
	S3Region     string

	// SES
	SESSenderEmail string
	SESRegion      string

	// Pine Labs
	PineLabsAPIURL     string
	PineLabsAPIKey     string
	PineLabsMerchantID string

	// JWT
	JWTSecret      string
	JWTExpiryHours int
}

func Load() *Config {
	return &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		MockMode:    getEnvBool("MOCK_MODE", true),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "supplierpay"),
		DBPassword: getEnv("DB_PASSWORD", "supplierpay_dev"),
		DBName:     getEnv("DB_NAME", "supplierpay"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		AWSRegion: getEnv("AWS_REGION", "ap-south-1"),

		BedrockModelID:     getEnv("BEDROCK_MODEL_ID", "anthropic.claude-3-5-sonnet-20241022-v2:0"),
		BedrockMaxTokens:   getEnvInt("BEDROCK_MAX_TOKENS", 4096),
		BedrockBearerToken: getEnv("AWS_BEARER_TOKEN_BEDROCK", ""),

		S3BucketName: getEnv("S3_BUCKET_NAME", "supplierpay-invoices"),
		S3Region:     getEnv("S3_REGION", "ap-south-1"),

		SESSenderEmail: getEnv("SES_SENDER_EMAIL", "noreply@supplierpay.ai"),
		SESRegion:      getEnv("SES_REGION", "ap-south-1"),

		PineLabsAPIURL:     getEnv("PINELABS_API_URL", "https://sandbox.pinelabs.com/api/v1"),
		PineLabsAPIKey:     getEnv("PINELABS_API_KEY", ""),
		PineLabsMerchantID: getEnv("PINELABS_MERCHANT_ID", ""),

		JWTSecret:      getEnv("JWT_SECRET", "supplierpay-hackathon-secret"),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return i
}
