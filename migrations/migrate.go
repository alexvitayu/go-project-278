package migrations

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Up применяет все pending миграции
func Up(pool *pgxpool.Pool) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(MigrationsFS)

	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	log.Println("✅ Migrations applied successfully")
	return nil
}

//// Down откатывает последнюю миграцию
//func Down(pool *pgxpool.Pool) error {
//	sqlDB := stdlib.OpenDBFromPool(pool)
//	defer sqlDB.Close()
//
//	if err := goose.Down(sqlDB, "."); err != nil {
//		return fmt.Errorf("migrate down: %w", err)
//	}
//
//	log.Println("✅ Migration rolled back successfully")
//	return nil
//}
//
//// Status показывает статус миграций
//func Status(pool *pgxpool.Pool) error {
//	sqlDB := stdlib.OpenDBFromPool(pool)
//	defer sqlDB.Close()
//
//	if err := goose.Status(sqlDB, "."); err != nil {
//		return fmt.Errorf("status: %w", err)
//	}
//
//	return nil
//}
//
//// Version возвращает текущую версию БД
//func Version(pool *pgxpool.Pool) (int64, error) {
//	sqlDB := stdlib.OpenDBFromPool(pool)
//	defer sqlDB.Close()
//
//	version, err := goose.GetDBVersion(sqlDB)
//	if err != nil {
//		return 0, fmt.Errorf("get version: %w", err)
//	}
//
//	return version, nil
//}
