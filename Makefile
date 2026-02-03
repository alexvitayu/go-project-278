ifeq (,$(wildcard .env.development))
    ifneq (,$(wildcard .env.example))
        $(info 📝 Creating .env.development from .env.example...)
        $(shell cp .env.example .env.development)
    endif
endif

-include .env.development
# ====================
# BUILD
# ====================

# Собирает бинарный файл в bin/lshortener
build:
	go build -o bin/lshortener ./cmd/lshortener

# Устанавливает собранный бинарник в GOBIN, чтобы его можно было запускать из любого места.
install: build
	go install ./cmd/lshortener

# Запуск линтера (использует .golangci.yml)
lint:
	golangci-lint run

# Запуск тестов
test:
	go test ./... -v -race -cover

# ====================
# LAUNCH FRONTEND + BACKEND
# ====================
run:
	concurrently 'npx start-hexlet-url-shortener-frontend' 'go run ./cmd/lshortener'

# ====================
# DEVELOPMENT (основные команды)
# ====================

# Полный запуск разработки
dev: dev-db-up dev-migrate-up dev-app-run

# Запуск только БД
dev-db-up:
	docker compose -f docker-compose-dev.yaml up -d
	@echo "✅ PostgreSQL запущена на $(DB_HOST):$(DB_PORT)"

# Остановка БД
dev-db-down:
	docker compose -f docker-compose-dev.yaml down

# Статус БД
dev-db-status:
	docker compose -f docker-compose-dev.yaml ps

# Запуск приложения (локально)
dev-app-run:
	APP_ENV=development go run ./cmd/lshortener

# Миграции для разработки

dev-migrate-up:
	export GOOSE_DRIVER=$(GOOSE_DRIVER) && \
    export GOOSE_DBSTRING=$(DATABASE_URL) && \
    goose -dir ./db/migrations up

dev-migrate-down:
	export GOOSE_DRIVER=$(GOOSE_DRIVER) && \
    export GOOSE_DBSTRING=$(DATABASE_URL) && \
    goose -dir ./db/migrations down

dev-migrate-status:
	export GOOSE_DRIVER=$(GOOSE_DRIVER) && \
    export GOOSE_DBSTRING=$(DATABASE_URL) && \
    goose -dir ./db/migrations status

# Полная очистка и пересоздание
dev-db-clean:
	@echo "Очистка старой БД..."
	docker compose -f docker-compose-dev.yaml down
	docker volume rm postgres-dev_data 2>/dev/null || true
	@echo "Создание новой БД..."
	$(MAKE) dev-db-up
	@sleep 3
	$(MAKE) dev-migrate-up

# Запуск генерации кода на основе файла sqlc.yaml
dev-sqlc:
	sqlc generate

dev-check-env:
	$(call load_env)
	@echo "📋 Проверка переменных окружения:"
	@echo "DB_HOST=$(DB_HOST)"
	@echo "DB_PORT=$(DB_PORT)"
	@echo "DB_NAME=$(DB_NAME)"
	@echo "DATABASE_URL=$(shell echo '$(DATABASE_URL)' | sed 's/:.*@/:****@/')"