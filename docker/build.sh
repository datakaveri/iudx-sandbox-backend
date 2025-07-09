#!/bin/bash

# To be executed from project root
docker build -t private-registry.iudx.org.in/sandbox-backend-prod:latest -f docker/prod.dockerfile .
docker build -t private-registry.iudx.org.in/sandbox-backend-dev:latest -f docker/dev.dockerfile .
docker build -t private-registry.iudx.org.in/sandbox-backend-test:latest -f docker/test.dockerfile .