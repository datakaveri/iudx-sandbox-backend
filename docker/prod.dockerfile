FROM golang:1.17.5-alpine AS build

RUN apk add --no-cache git
WORKDIR /app
ARG API_PORT
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main cmd/api/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/main /usr/bin/
ENTRYPOINT [ "main" ]