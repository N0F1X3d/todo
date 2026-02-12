# Build stage
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Копируем pkg (общий код)
COPY pkg/ ./pkg/

# Копируем api-service
COPY api-service/ ./api-service/

# Переходим в директорию api-service
WORKDIR /app/api-service

# Скачиваем зависимости
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o api-service ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/api-service/api-service .

CMD ["./api-service"]