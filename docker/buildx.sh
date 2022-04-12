#!/bin/sh

# docker buildx create --use

BUILD_COMMIT=$(git log -1 --format=%H) BUILD_DATE=$(date) docker buildx bake -f $(dirname $0)/docker-compose.build.yml --pull $@
