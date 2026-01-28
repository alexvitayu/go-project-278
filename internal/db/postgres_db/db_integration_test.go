// db_integration_test.go содержит только TestMain и общие утилиты
package postgres_db

import (
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
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const BASE_URL = "http://localhost:8080"

var (
	pool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

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
	pool, err = NewTestPgxPool(ctx, dsn)
	if err != nil {
		log.Fatalf("creation pool: %v", err)
	}

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

func NewTestPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := p.Ping(ctx); err != nil {
		return nil, fmt.Errorf("fail to ping database: %w", err)
	}

	return p, nil
}

func withTx(t *testing.T, fn func(ctx context.Context, q *Queries)) {
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

	qtx := New(tx) // все вызовы sqlc пойдут внутри этой транзакции
	fn(ctx, qtx)
}

func CreateTestLinks(t *testing.T, ctx context.Context, q *Queries, baseURL string) ([]*CreateLinkRow, error) {
	t.Helper()
	params := []CreateLinkParams{
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
	links := make([]*CreateLinkRow, 0, len(params))
	for _, v := range params {
		row, err := q.CreateLink(ctx, v)
		if err != nil {
			return nil, err
		}
		links = append(links, &row)
	}
	return links, nil
}
