#!/bin/bash

docker run --rm -it \
  --env-file .env \
  -p 3000:3000 \
  --name cano-collector \
  kubecano/cano-collector:"${1:-latest}"
