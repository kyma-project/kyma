#!/usr/bin/env bash

export DOCKER_TAG="PR-14743"
export DOCKER_PUSH_REPOSITORY="eu.gcr.io/kyma-project"
make release
