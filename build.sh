#!/usr/bin/env bash
set -e 
echo "building linux executable..."
GOOS=linux go build 
echo "building Docker container image..."
docker build -t davestearns/userservice .
echo "cleaning up..."
go clean 
