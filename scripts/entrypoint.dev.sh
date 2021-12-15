#!/bin/bash
set -e

echo "############ Downloading CompileDaemon ############"
# disable go modules to avoid this package from getting into go.mod
GO111MODULE=off go get github.com/githubnemo/CompileDaemon

echo "############ Starting Deamon ############"
CompileDaemon --build="go build -o main cmd/api/main.go" --command=./main
