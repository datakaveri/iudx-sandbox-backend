FROM golang:1.24.4-alpine

RUN apk add --no-cache build-base
WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .