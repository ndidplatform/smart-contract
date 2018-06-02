#!/bin/sh

BUILD_COMMIT=$(git log -1 --format=%H) BUILD_DATE=$(date) docker-compose -f $(dirname $0)/docker-compose.build.yml build $@
