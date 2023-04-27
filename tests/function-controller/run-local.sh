#!/bin/bash
kubectl proxy &
APP_TEST_KUBECTL_PROXY_ENABLED=true go run ./cmd/main.go $1
EXIT_CODE=$?
pkill kubectl
exit ${EXIT_CODE}
