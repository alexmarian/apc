#!/bin/bash
set -e

KEEP=0
BUILD_VITE=0
BUILD_API=0

# Parse arguments
for arg in "$@"; do
  case $arg in
    --keep) KEEP=1 ;;
    --vite) BUILD_VITE=1 ;;
    --api)  BUILD_API=1 ;;
  esac
done

# Default: build both if neither specified
if [[ $BUILD_VITE -eq 0 && $BUILD_API -eq 0 ]]; then
  BUILD_VITE=1
  BUILD_API=1
fi

trap '[[ $KEEP -eq 0 ]] && rm -rf disttmp; docker rm -f transfer_container 2>/dev/null || true' EXIT

rm -rf disttmp
mkdir -p disttmp/static

if [[ $BUILD_VITE -eq 1 ]]; then
  cd ../ui && npm install && npm run build && cd ../api
  cp -R ../ui/dist/* disttmp/static/
fi

if [[ $BUILD_API -eq 1 ]]; then
  docker build --target export-stage -f dist/Dockerfile --output disttmp .
fi
if [[ -f ../dist.zip ]]; then
  rm ../dist.zip
fi
cd disttmp
zip -r ../dist.zip .
cd ..
if [[ $KEEP -eq 0 ]]; then
  rm -rf disttmp
fi