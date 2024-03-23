#!/bin/bash

# Retrieve version from the file or environment variable
HEALTHCHECK_PROXY_VERSION=$(cat version 2>/dev/null || echo "${HEALTHCHECK_PROXY_VERSION}")

# Check if version is set
if [ -z "${HEALTHCHECK_PROXY_VERSION}" ]; then
  echo "HEALTHCHECK_PROXY_VERSION is not set and auto-detection failed."
  exit 1
fi

echo "Building healthcheck-proxy version ${HEALTHCHECK_PROXY_VERSION} for amd64/arm64"

# Create and use buildx builder
docker buildx create --use --name=healthcheck-proxy

# Authenticate with GitHub Container Registry using the token from GitHub Secrets
echo "${GHCR_TOKEN}" | docker login ghcr.io -u "${GITHUB_ACTOR}" --password-stdin

# Build and push the Docker image
docker buildx build --platform linux/amd64,linux/arm64 . \
  --tag "ghcr.io/${GITHUB_REPOSITORY}/healthcheck-proxy:${HEALTHCHECK_PROXY_VERSION}" \
  --tag "ghcr.io/${GITHUB_REPOSITORY}/healthcheck-proxy:latest" \
  --push

# Check if build and push were successful
if [ $? -eq 0 ]; then
  echo "Build and push completed successfully."
else
  echo "Build and push failed."
  exit 1
fi