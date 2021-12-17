FROM golang:1.17.5-alpine

RUN apk add --no-cache build-base
WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .