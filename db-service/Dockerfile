# ---------- build stage ----------
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o db-service ./cmd/db-service

# ---------- runtime stage ----------
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/db-service .

EXPOSE 50051

CMD ["./db-service"]
