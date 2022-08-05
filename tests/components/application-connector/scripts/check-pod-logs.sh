#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)
JOB_NAME=$1

if [ $# -ne 1 ]; then
  echo "Usage: check-pod-logs.sh <job name>"
  exit 1
fi

podName=$(kubectl get pods -n $NAMESPACE --selector=job-name=$1 --output=jsonpath='{.items[*].metadata.name}')

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs $podName -f | tee /dev/stderr | $GOPATH/bin/go-junit-report -subtest-mode exclude-parents -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs $podName application-gateway-test
fi
