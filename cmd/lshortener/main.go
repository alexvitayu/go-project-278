package main

import (
	"code/internal/db/postgres_db"
	"code/internal/handlers"
	"code/internal/service"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // blank identifier означает, что пакет импортирован без прямого использования в коде
	"github.com/joho/godotenv"
)

const (
	maxOpenConn = 10
	maxIdleConn = 5
	maxLifetime = 30 * time.Minute
)

func main() {
	// Load загружает переменные из файла .env в окружение текущего процесса.
	// После вызова доступны через os.Getenv().
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env." + env

	if err := godotenv.Load(envFile); err != nil {
		log.Fatal("Error loading .env file")
	}

	DSN := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSL_MODE"),
	)

	log.Printf("Connecting to: %s", DSN)
	// Создаём контекст с таймаутом. Если база "зависла", приложение не будет ждать бесконечно.
	ctx, cancel := context.WithTimeout(context.Background(), 1800*time.Second)
	defer cancel()

	// Создаём подключение
	pool := NewPgxPool(ctx, DSN)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()
	//
	//// Настраиваем пул соединений
	//db.SetMaxOpenConns(maxOpenConn)    // Если запросов больше, лишние будут ждать свободный слот.
	//db.SetMaxIdleConns(maxIdleConn)    // Полезно, когда нагрузка скачет: свободные соединения не закрываются сразу.
	//db.SetConnMaxLifetime(maxLifetime) // Старые соединения будут обновляться — это защищает от зависших сессий.
	//
	//if err := db.PingContext(ctx); err != nil {
	//	log.Fatal("database unreachable:", err)
	//}
	//
	//log.Println("Successfully connected to PostgreSQL DB!")

	queries := postgres_db.New(pool)

	service := service.NewLinkService(queries)

	router := handlers.SetupRouter(ctx, service)

	router.Use(gin.Recovery())

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
}

func NewPgxPool(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	return pool
}
