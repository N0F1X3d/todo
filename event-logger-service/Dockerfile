# ---------- build stage ----------
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY pkg/ ./pkg/
COPY event-logger-service/ ./event-logger-service/

WORKDIR /app/event-logger-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o event-logger-service ./cmd

# ---------- runtime stage ----------
FROM alpine:3.19

WORKDIR /app

# создаём папку для логов
RUN mkdir /logs

COPY --from=builder /app/event-logger-service/event-logger-service .

CMD ["./event-logger-service"]
