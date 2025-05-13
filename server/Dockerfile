
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o yartime-server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/yartime-server .
COPY config/config.json ./config/config.json
RUN chmod 644 ./config/config.json
EXPOSE 8080
CMD ["./yartime-server"]