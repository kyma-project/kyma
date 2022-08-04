#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)

if ([[ ${EXPORT_RESULT} == true ]]); then
<<<<<<< HEAD
<<<<<<< HEAD
	kubectl -n $NAMESPACE logs -f job/application-gateway-test | tee /dev/stderr | $GOPATH/bin/go-junit-report -set-exit-code > junit-report.xml
=======
	kubectl -n $NAMESPACE logs $podName -f | tee /dev/stderr | $GOPATH/bin/go-junit-report -subtest-mode -set-exit-code > junit-report.xml
>>>>>>> 8696c3a16 (Ignore results of subtest parent tests in JUnit)
=======
	kubectl -n $NAMESPACE logs $podName -f | tee /dev/stderr | $GOPATH/bin/go-junit-report/v2 -subtest-mode -set-exit-code > junit-report.xml
>>>>>>> fc4b57f50 (Change the using version of junit in makefile)
else
	kubectl -n $NAMESPACE logs -f job/application-gateway-test
fi
