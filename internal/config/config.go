package config

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	APPEnv       string
	ServerPort   string
	BaseURL      string
	DBConfig     DBConfig
	PoolConfig   PoolConfig
	GooseConfig  GooseConfig
	SentryConfig SentryConfig
}

type DBConfig struct {
	DATABASE_URL string
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	DBSSLMode    string
}

type PoolConfig struct {
	DBMaxConns        string
	DBMaxIdleConns    string
	DBConnMaxLifetime string
}

type GooseConfig struct {
	GooseDBString string
	GooseDriver   string
}

type SentryConfig struct {
	SentryDSN string
}

func Load() (*AppConfig, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Загружаем env файл для конкретной среды
	if env == "development" {
		if _, err := os.Stat(".env.development"); err == nil {
			if err := godotenv.Load(".env.development"); err != nil {
				return nil, fmt.Errorf("loading %s: %w", ".env.development", err)
			}
		}
	}

	dbConfig := DBConfig{
		DATABASE_URL: getEnv("DATABASE_URL", ""),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBName:       getEnv("DB_NAME", ""),
		DBUser:       getEnv("DB_USER", ""),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DBSSLMode:    getEnv("DB_SSL_MODE", "disable"),
	}

	// Получаем значения из переменных окружения
	config := &AppConfig{
		APPEnv:     env,
		ServerPort: getEnv("PORT", "8080"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),
		DBConfig:   dbConfig,
		PoolConfig: PoolConfig{
			DBMaxConns:        getEnv("DB_MAX_CONNS", "10"),
			DBMaxIdleConns:    getEnv("DB_MAX_IDLE_CONNS", "5"),
			DBConnMaxLifetime: getEnv("DB_CONN_MAX_LIFETIME", "30m"),
		},
		GooseConfig: GooseConfig{
			GooseDBString: dbConfig.DATABASE_URL,
			GooseDriver:   getEnv("GOOSE_DRIVER", "postgres"),
		},
		SentryConfig: SentryConfig{
			SentryDSN: getEnv("SENTRY_DSN", ""),
		},
	}

	return config, nil
}

// getEnv получает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func SetCORSConfig(env string) cors.Config {
	if env == "production" {
		return cors.Config{
			AllowOrigins:     []string{"https://go-project-278.onrender.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Range"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}
	}
	return cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}
}
