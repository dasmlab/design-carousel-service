#!/bin/bash
# runme_local.sh

# This runs very privileged (host network mode, privileged, etc) and its not meant for production. Use the K8s Envelope for more 'production' ready packaging and runtime goodness


app=design-carousel-service
version=latest

docker stop ${app}-instance
docker rm  ${app}-instance

if [[ -z "$app" || -z "$version" ]]; then
  echo "ERROR: app or version variable is not set!"
  exit 1
fi

echo "Running: docker run ... $app:$version"
docker run -d -p 10022:10022 -p 9222:9222 --name ${app}-instance \
  --restart=always \
  ${app}:${version}

