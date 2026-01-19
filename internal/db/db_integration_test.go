// db_integration_test.go содержит только TestMain и общие утилиты
package db_test

import (
	"code/internal/config"
	"code/internal/db/postgres_db"
	"code/migrations"
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pool       *pgxpool.Pool
	testConfig *config.AppConfig
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Устанавливаем APP_ENV=test, чтобы брать данные из .env.test
	os.Setenv("APP_ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load test config: %v", err)
	}
	testConfig = cfg

	fmt.Printf("Loaded config: APP_ENV=%s, DB_HOST=%s\n, BASE_URL=%s\n",
		testConfig.APPEnv, testConfig.DBConfig.DBHost, testConfig.BaseURL)

	// Запуск PostgreSQL контейнера
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithSQLDriver("pgx/v5"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		tc.WithAdditionalWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start container: %v", err)
	}
	defer func() { _ = container.Terminate(ctx) }()

	// Получение DSN
	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432/tcp")
	dsn := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=testdb sslmode=disable",
		host,
		port.Port(),
	)
	//Создание пула соединений
	pool = NewTestPgxPool(ctx, dsn)

	defer pool.Close()

	// Конвертируем pgxpool.Pool в *sql.DB
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	// Применение миграций
	goose.SetBaseFS(migrations.MigrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	//Запуск тестов
	code := m.Run()
	os.Exit(code)
}

func NewTestPgxPool(ctx context.Context, dsn string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	return pool
}

func withTx(t *testing.T, fn func(ctx context.Context, q *postgres_db.Queries)) {
	t.Helper()

	// Базовый контекст — из теста.
	// Если нужно, можно поверх навесить timeout.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	// Любой тест либо сам закоммитит транзакцию, либо она откатится в конце.
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	// Сброс последовательности перед тестом
	_, err = tx.Exec(ctx, `TRUNCATE TABLE links RESTART IDENTITY CASCADE`)
	require.NoError(t, err)

	qtx := postgres_db.New(tx) // все вызовы sqlc пойдут внутри этой транзакции
	fn(ctx, qtx)
}

func CreateTestLinks(t *testing.T, ctx context.Context, q *postgres_db.Queries, baseURL string) ([]*postgres_db.CreateLinkRow, error) {
	t.Helper()
	params := []postgres_db.CreateLinkParams{
		{
			OriginalUrl: "https://example1.net/very-very-long-short-name?with=queries",
			ShortName:   "test-short1",
			ShortUrl:    baseURL + "/test-short1"},
		{
			OriginalUrl: "https://example2.net/very-very-long-short-name?with=queries",
			ShortName:   "test-short2",
			ShortUrl:    baseURL + "/test-short2"},
		{
			OriginalUrl: "https://example3.net/very-very-long-short-name?with=queries",
			ShortName:   "test-short3",
			ShortUrl:    baseURL + "/test-short3"},
	}
	links := make([]*postgres_db.CreateLinkRow, 0, len(params))
	for _, v := range params {
		row, err := q.CreateLink(ctx, v)
		if err != nil {
			return nil, err
		}
		links = append(links, &row)
	}
	return links, nil
}
