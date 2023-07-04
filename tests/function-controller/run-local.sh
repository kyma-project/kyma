#!/bin/bash
kubectl proxy &
KUBECTL_PID=$!

APP_TEST_KUBECTL_PROXY_ENABLED=true go run ./cmd/main.go $1
EXIT_CODE=$?
kill $KUBECTL_PID
exit ${EXIT_CODE}
