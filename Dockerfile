# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server-management ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server-management .
COPY ./internal/config.go ./internal/config.go
EXPOSE 8080
CMD ["./server-management"]
