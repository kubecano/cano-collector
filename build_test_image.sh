#!/bin/bash

TAG=$(git rev-parse --short HEAD)
docker buildx build \
 --tag kubecano/cano-collector:${TAG} \
 --tag kubecano/cano-collector:latest \
 .