# ---------- build stage ----------
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Копируем pkg (общий код)
COPY pkg/ ./pkg/

# Копируем db-service
COPY db-service/ ./db-service/

# Переходим в директорию db-service
WORKDIR /app/db-service

# Скачиваем зависимости
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o db-service ./cmd/db-service

# ---------- runtime stage ----------
FROM alpine:3.19

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/db-service/db-service .

EXPOSE 50051

CMD ["./db-service"]