# Этап сборки
FROM golang:1.24-bullseye AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    librdkafka-dev \
    pkg-config \
    git \
    && rm -rf /var/lib/apt/lists/*

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарники с CGO
RUN CGO_ENABLED=1 go build -o post-service ./cmd/app
RUN go build -o migrator ./cmd/migrator

# Финальный образ на Debian
FROM debian:bullseye-slim

WORKDIR /app

# Устанавливаем runtime-зависимости
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    librdkafka1 \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Копируем всё нужное
COPY --from=builder /app/post-service .
COPY --from=builder /app/migrator .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config ./config

EXPOSE 8000

CMD ["sh", "-c", "./migrator --migrations-path ./migrations && ./post-service"]
