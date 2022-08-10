#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)
JOB_NAME="$1"

podName=$(kubectl get pods -n $NAMESPACE --selector=job-name=$JOB_NAME --output=jsonpath='{.items[*].metadata.name}')

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs $podName -f | tee /dev/stderr | $GOPATH/bin/go-junit-report -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs $podName application-gateway-test
fi