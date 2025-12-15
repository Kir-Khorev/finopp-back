package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Environment   string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	JWTSecret     string
	GroqAPIKey    string
	FixerAPIKey   string
}

func Load() *Config {
	// Load .env file if exists
	_ = godotenv.Load()

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENV", "development"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "finopp"),
		DBPassword:    getEnv("DB_PASSWORD", "finopp_pass"),
		DBName:        getEnv("DB_NAME", "finopp_db"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		GroqAPIKey:    getEnv("GROQ_API_KEY", ""),
		FixerAPIKey:   getEnv("FIXER_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue == "" {
		log.Printf("Warning: %s is not set", key)
	}
	return defaultValue
}

