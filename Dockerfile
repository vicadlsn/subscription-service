FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o subscription-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/subscription-service .

EXPOSE 8081
ENTRYPOINT ["./subscription-service"]