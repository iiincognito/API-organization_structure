FROM golang:1.25.5-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем goose в builder-стадии
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Финальный лёгкий образ
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Копируем goose из builder
COPY --from=builder /go/bin/goose /usr/local/bin/goose

WORKDIR /app
COPY --from=builder /app/api .
COPY migrations ./migrations

CMD ["./api"]