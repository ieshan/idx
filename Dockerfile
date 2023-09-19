FROM golang:1.21.1-bookworm

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
