FROM golang:1.17.5

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .