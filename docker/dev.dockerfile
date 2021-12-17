FROM golang:1.17.5-alpine

RUN apk add --no-cache git
WORKDIR /app

ARG API_PORT

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
