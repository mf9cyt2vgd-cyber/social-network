# Этап сборки
FROM golang:latest AS builder

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
RUN CGO_ENABLED=1 go build -o notification-service ./cmd/app

# Финальный образ на Debian совместимый с C-библиотеками для Kafka
FROM debian:bookworm-slim

WORKDIR /app

# Устанавливаем runtime-зависимости
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

# Копируем всё нужное
COPY --from=builder /app/notification-service .
COPY --from=builder /app/config ./config

EXPOSE 8001

CMD ["./notification-service"]