FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates mysql-client openssl
WORKDIR /app
COPY --from=builder /build/server .
COPY --from=builder /build/migrations ./migrations
RUN mkdir -p /app/storage/images /app/storage/quarantine /app/storage/backups /app/certs
RUN openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /app/certs/server.key -out /app/certs/server.crt \
    -subj "/C=US/ST=Local/L=Local/O=LocalInsights/CN=localhost" 2>/dev/null
EXPOSE 8443
CMD ["./server"]
