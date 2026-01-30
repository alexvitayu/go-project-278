# 1) Build frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /build/frontend

COPY package*.json ./

RUN --mount=type=cache,target=/root/.npm \
  npm ci npm ci --prefer-offline --no-audit

# 2) Build backend
FROM golang:1.25-alpine AS backend-builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git

RUN mkdir -p /build/code

# Рабочая директория
WORKDIR /build/code

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
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app ./cmd/lshortener/

# 3) Runtime
FROM alpine:3.22

# Устанавливаем Caddy, bash, пакеты для PostgreQSL + зависимости
RUN apk add --no-cache \
    bash \
    postgresql-client \
    ca-certificates \
    caddy \
    tzdata

WORKDIR /app

## Копируем бинарник
COPY --from=backend-builder /build/app /app/bin/app
COPY --from=frontend-builder \
  /build/frontend/node_modules/@hexlet/project-url-shortener-frontend/dist \
  /app/public

## Копируем миграции
COPY --from=backend-builder /build/code/db/migrations /app/db/migrations

## Копируем goose
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

## Копируем скрипт запуска
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 80

CMD ["/app/bin/run.sh"]