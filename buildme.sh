#!/bin/bash

set -e

app=design-carousel-service
version=latest
ARCH="amd64"

docker build -t ${app}:${version} \
  --build-arg ARCH=$ARCH \
  --build-arg goproxy=https://proxy.golang.org \
  -f Dockerfile .

