#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)
JOB_NAME="$1"

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs -f job/$JOB_NAME | tee /dev/stderr | $GOPATH/bin/go-junit-report -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs -f job/$JOB_NAME
fi
