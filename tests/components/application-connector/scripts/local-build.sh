#!/usr/bin/env bash

if [ $# -ne 2 ]; then
  echo "Usage: local_build.sh <DOCKER_TAG> <DOCKER_PUSH_REPOSITORY>"
  exit 1
fi

export DOCKER_TAG=$1
export DOCKER_PUSH_REPOSITORY=$2
make release
