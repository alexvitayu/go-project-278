#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting service"

# Проверяем переменные окружения
if [ -z "${DATABASE_URL:-}" ]; then
    echo "Error: DATABASE_URL is not set"
    exit 1
fi

echo "[run.sh] Running DB migrations..."
goose -dir ./migrations postgres "${DATABASE_URL}" up

echo "[run.sh] Starting Go app..."
exec ./bin/lshortener