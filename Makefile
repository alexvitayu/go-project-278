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
	go test ./... -v -race
