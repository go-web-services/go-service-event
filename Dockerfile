# Stage 1: Build goose
FROM golang:1.26-alpine AS goose-builder
RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Stage 2: Build Stage
FROM golang:1.26-alpine AS builder

ARG GITHUB_TOKEN

ENV GO111MODULE=on \
    GOPRIVATE=github.com/Lomank123/*

RUN apk add --no-cache git ca-certificates tzdata && \
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o go-service-event ./cmd/app/main.go

# Stage 3: Run Stage
FROM alpine:3.20.3

WORKDIR /app

COPY --from=builder /app/go-service-event .
COPY --from=builder /app/migrations ./migrations
COPY --from=goose-builder /go/bin/goose /usr/local/bin/goose

CMD ["./go-service-event"]
