# Этап сборки
FROM golang:1.26-bookworm AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o post-service ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrator ./cmd/migrator

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/post-service .
COPY --from=builder /app/migrator .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config ./config

EXPOSE 8000

CMD ["sh", "-c", "./migrator --migrations-path ./migrations && ./post-service"]