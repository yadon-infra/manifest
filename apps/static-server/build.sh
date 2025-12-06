#!/bin/bash

set -e

REGISTRY="harbor.yadon3141.com"
PROJECT="infra"
IMAGE_NAME="static-server"
TAG="${1:-latest}"

FULL_IMAGE_NAME="${REGISTRY}/${PROJECT}/${IMAGE_NAME}:${TAG}"

echo "Building container image: ${FULL_IMAGE_NAME}"
sudo docker build -t "${FULL_IMAGE_NAME}" .

echo "Pushing container image to Harbor..."
echo "Please make sure you are logged in to Harbor:"
echo "  sudo docker login ${REGISTRY}"
echo ""
echo "Press Enter to continue or Ctrl+C to cancel..."
read

sudo docker push "${FULL_IMAGE_NAME}"

echo "Successfully pushed ${FULL_IMAGE_NAME}"
echo ""
echo "To deploy to Kubernetes, update the image tag in:"
echo "  k8s/application/static-server/components/deployment.yaml"
echo ""
echo "Then apply the configuration:"
echo "  kubectl apply -k k8s/application/static-server/"