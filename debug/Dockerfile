FROM golang:1.26-alpine AS goose-builder

RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM golang:1.26-alpine

ARG GITHUB_TOKEN

ENV GO111MODULE=on \
    GOPRIVATE=github.com/Lomank123/*

RUN apk add --no-cache git ca-certificates tzdata && \
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

COPY --from=goose-builder /go/bin/goose /go/bin/goose

WORKDIR /app
COPY . .

RUN go mod download
