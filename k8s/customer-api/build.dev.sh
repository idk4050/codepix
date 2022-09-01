#!/bin/sh

cd $(dirname "$0")

version="latest"
[ -n "$1" ] && version="$1"

image="registry.codepix.local/customer-api:$version"

echo "Building $image"
echo

podman build -t "$image" \
  -f api.Containerfile \
  --ignorefile api.containerignore \
  --target runner \
  --build-arg buildcache=$(go env GOCACHE) \
  --build-arg modcache=$(go env GOMODCACHE) ../../customer-api

echo
echo "Pushing $image"
echo

podman push --tls-verify=false "$image"
