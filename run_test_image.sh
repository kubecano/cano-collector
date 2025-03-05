#!/bin/bash

docker run --rm -it \
  --env-file .env \
  -p 8080:8080 \
  --name cano-collector \
  kubecano/cano-collector:"${1:-latest}"
