FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .
RUN swag init -g cmd/main.go
RUN go build -o subscription-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/subscription-service .
COPY --from=builder /app/docs ./docs

EXPOSE 8081
ENTRYPOINT ["./subscription-service"]