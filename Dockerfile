# Build backend
FROM golang:1.25-alpine AS backend-builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git

# Рабочая директория
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости с кэшированием
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Устанавливаем goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем весь код
COPY . .

# Собираем приложение
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bin/lshortener ./cmd/lshortener

# 3) Runtime
FROM alpine:3.22

WORKDIR /app

# Устанавливаем bash и необходимые пакеты, т.к. в run.sh используется bash
RUN apk add --no-cache bash postgresql-client ca-certificates

# Устанавливаем необходимые пакеты для работы с PostgreSQL
RUN apk add --no-cache postgresql-client ca-certificates

## Копируем бинарник
COPY --from=backend-builder /app/bin/lshortener ./bin/lshortener

## Копируем миграции
COPY --from=backend-builder /app/migrations ./migrations

## Копируем goose
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

## Копируем скрипт запуска
COPY bin/run.sh ./bin/run.sh
RUN chmod +x ./bin/run.sh

EXPOSE 8080

CMD ["./bin/run.sh"]