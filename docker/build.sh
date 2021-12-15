#!/bin/bash

# To be executed from project root
docker build -t iudx/sandbox-backend-prod:latest -f docker/prod.dockerfile .
docker build -t iudx/sandbox-backend-dev:latest -f docker/dev.dockerfile .
docker build -t iudx/sandbox-backend-test:latest -f docker/test.dockerfile .