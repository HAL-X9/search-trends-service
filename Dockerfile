# === ЭТАП 1: Сборка приложения ===
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Копируем всё содержимое проекта сразу
COPY . .

ARG SERVICE_NAME=search-trends

# Скачиваем и собираем в один приход, явно передавая переменную окружения
RUN GOPROXY=direct CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/${SERVICE_NAME}/main.go

# === ЭТАП 2: Финальный минимальный образ ===
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/main .
COPY --from=builder /build/config ./config

CMD ["./main"]