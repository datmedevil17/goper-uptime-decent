package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	RabbitMQURL        string
	PlatformPrivateKey string

	JWTSecret string
	Port      string
	HubURL    string
}

func Load() *Config {
	// Load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgresql://uptime_user:uptime_password@localhost:5432/uptime_db?sslmode=disable"),
		RabbitMQURL:        getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
		PlatformPrivateKey: getEnv("PLATFORM_PRIVATE_KEY", ""),

		JWTSecret: getEnv("JWT_SECRET", "super-secret-key-change-me"),
		Port:      getEnv("PORT", "8080"),
		HubURL:    getEnv("HUB_URL", "ws://localhost:8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
