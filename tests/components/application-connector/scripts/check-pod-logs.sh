#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)

podName=$(kubectl get pods -n $NAMESPACE --selector=job-name=application-gateway-test --output=jsonpath='{.items[*].metadata.name}')

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs $podName -f | tee /dev/stderr | $GOPATH/bin/go-junit-report -subtest-mode exclude-parents -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs $podName application-gateway-test
fi