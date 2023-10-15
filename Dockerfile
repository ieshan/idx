FROM golang:1.21.3-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
