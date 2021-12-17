#!/bin/bash
set -e

if [ ${TEST_COVERAGE:-false} == true ]
then 
    go test -coverprofile=coverage.out ./...
else 
    go test -v ./...
fi
