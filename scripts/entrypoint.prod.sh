#!/bin/bash
set -e

echo "############ Building application ############"
go build -o main cmd/api/main.go

./main
