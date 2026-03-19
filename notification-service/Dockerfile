FROM golang:1.26-alpine AS builder


RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notification-service ./cmd/app

FROM gcr.io/distroless/base-debian11

WORKDIR /app

# Копируем только бинарник и конфиг
COPY --from=builder /app/notification-service .
COPY --from=builder /app/config ./config

# Указываем порт
EXPOSE 8001

# Запуск приложения
CMD ["./notification-service"]