#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

mockery -case underscore -name DiscoveryInterface -dir ./vendor/k8s.io/client-go/discovery/ -output ./internal/domain/k8s/automock -outpkg automock
