# ---- Build Stage ----
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 go build -o novel-service .

# ---- Runtime Stage ----
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app/backend

COPY --from=builder /app/novel-service .
COPY backend/config.yaml ./config.yaml
COPY frontend/ ../frontend/

EXPOSE 8001

CMD ["./novel-service"]
