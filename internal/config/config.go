package config

import (
	"fmt"
	"os"

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
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
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
	envFile := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("loading %s: %w", envFile, err)
		}
	}

	dbConfig := DBConfig{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnvRequired("DB_NAME"),
		DBUser:     getEnvRequired("DB_USER"),
		DBPassword: getEnvRequired("DB_PASSWORD"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
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
			GooseDBString: dbConfig.ConnectionString(),
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

// getEnvRequired получает обязательную переменную окружения
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

func (c *DBConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}
