# ====== Этап 1: Сборка ======
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0 — статический бинарник
# -ldflags="-s -w" — убираем отладочную инфу (экономия ~30% размера)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o /bot ./main.go

# ====== Этап 2: Минимальный runtime ======
FROM scratch

# Корневые сертификаты — обязательны для HTTPS к Telegram API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Часовая зона (опционально, для красивых логов)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /bot /bot

ENV TZ=UTC
EXPOSE 8080

CMD ["/bot"]