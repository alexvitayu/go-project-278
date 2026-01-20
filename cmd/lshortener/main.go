package main

import (
	"code/internal/config"
	"code/internal/db/postgres_db"
	"code/internal/handlers"
	"code/internal/service"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // blank identifier означает, что пакет импортирован без прямого использования в коде
)

const DefaultTimeout = 30 * time.Minute

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Создаём контекст с таймаутом. Если база "зависла", приложение не будет ждать бесконечно.
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	pool, err := NewPgxPool(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("✅ Database connected")

	queries := postgres_db.New(pool)

	service := service.NewLinkService(queries, cfg)

	router := handlers.SetupRouter()

	router.Use(gin.Recovery())

	handlers := handlers.NewHandler(service)

	router.GET("/", handlers.HomePage)
	router.POST("/api/links", handlers.CreateLink)
	router.GET("/api/links", handlers.GetLinks)
	router.GET("/api/links/:id", handlers.GetLinkByID)
	router.PUT("/api/links/:id", handlers.UpdateLinkByID)
	router.DELETE("/api/links/:id", handlers.DeleteLinkByID)

	port := cfg.ServerPort

	if err := router.Run(fmt.Sprintf("0.0.0.0:%v", port)); err != nil {
		log.Fatalf("не удалось запустить сервер на порту %v: %v", port, err)
	}
}

func NewPgxPool(ctx context.Context, cfg *config.AppConfig) (*pgxpool.Pool, error) {
	// Парсим конфиг из DSN
	conf, err := pgxpool.ParseConfig(cfg.DBConfig.DATABASE_URL)
	if err != nil {
		return nil, fmt.Errorf("parse conf: %w", err)
	}
	maxConns, err := strconv.ParseInt(cfg.PoolConfig.DBMaxConns, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parseInt: %w", err)
	}
	minConns, err := strconv.ParseInt(cfg.PoolConfig.DBMaxIdleConns, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parseInt: %w", err)
	}
	maxLifetime, _ := time.ParseDuration(cfg.PoolConfig.DBConnMaxLifetime)
	conf.MaxConns = int32(maxConns)
	conf.MinConns = int32(minConns)
	conf.MaxConnIdleTime = maxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}
