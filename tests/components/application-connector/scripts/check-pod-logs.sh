#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)
JOB_NAME=$1

if [ $# -ne 1 ]; then
  echo "Usage: check-pod-logs.sh <job name>"
  exit 1
fi

POD_NAME=$(kubectl get pods -n $NAMESPACE --selector=job-name=$JOB_NAME --output=jsonpath='{.items[*].metadata.name}')

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs $POD_NAME -f | tee /dev/stderr | $GOPATH/bin/go-junit-report -subtest-mode exclude-parents -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs $POD_NAME application-gateway-test
fi
