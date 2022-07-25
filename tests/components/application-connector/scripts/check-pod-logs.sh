#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs -f job/application-gateway-test | tee /dev/stderr | $GOPATH/bin/go-junit-report -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs -f job/application-gateway-test
fi
