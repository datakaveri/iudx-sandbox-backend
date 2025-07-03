FROM golang:1.24.4-alpine

RUN apk add --no-cache git
WORKDIR /app

ARG API_PORT

COPY ./go.mod ./go.sum ./
RUN go mod download

ENV TZ="Asia/Kolkata"

COPY . .
