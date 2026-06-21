# syntax=docker/dockerfile:1

# ---------- Этап 1: сборка ----------
# go-sqlite3 использует cgo, поэтому собираем на полноценном образе с gcc.
FROM golang:1.22-bookworm AS build

WORKDIR /src

# Сначала копируем только go.mod/go.sum — так Docker закэширует слой
# с зависимостями и не будет перекачивать их при каждом изменении кода.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=1 обязателен для mattn/go-sqlite3.
RUN CGO_ENABLED=1 GOOS=linux go build -o /out/examplan .

# ---------- Этап 2: финальный лёгкий образ ----------
FROM debian:bookworm-slim

# ca-certificates нужен, если приложение когда-либо будет делать исходящие
# HTTPS-запросы; libc уже есть в базовом образе debian-slim.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/examplan ./examplan
COPY templates ./templates
COPY static ./static

# На бесплатном тарифе Render файловая система эфемерна — это ожидаемо
# для SQLite в этом проекте (история сбрасывается при рестарте/сне).
ENV EXAMPLAN_DB=/app/examplan.db

EXPOSE 10000

CMD ["./examplan"]
