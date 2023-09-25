#!/usr/bin/env bash
NAMESPACE=test
GOPATH=$(go env GOPATH)
JOB_NAME=$1

if [ $# -ne 1 ]; then
  echo "Usage: check-pod-logs.sh <job name>"
  exit 1
fi


time_so_far=0 # we've already waited that many seconds
sleep_sec=5 # wait between checks
wait_timeout=900 # 15min -> 900sec
retval_complete=1
retval_failed=1
while [[ $retval_complete -ne 0 ]] && [[ $retval_failed -ne 0 ]] && [[ $time_so_far -le $wait_timeout ]]; do
  sleep $sleep_sec
  time_so_far=$((time_so_far+sleep_sec))

  output=$(kubectl wait --for=condition=failed -n $NAMESPACE job/$JOB_NAME --timeout=0 2>&1)
  retval_failed=$?
  output=$(kubectl wait --for=condition=complete -n $NAMESPACE job/$JOB_NAME --timeout=0 2>&1)
  retval_complete=$?
done

if ([[ ${EXPORT_RESULT} == true ]]); then
	kubectl -n $NAMESPACE logs -f job/$JOB_NAME | tee /dev/stderr | $GOPATH/bin/go-junit-report -subtest-mode exclude-parents -set-exit-code > junit-report.xml
else
	kubectl -n $NAMESPACE logs -f job/$JOB_NAME
fi

if [ $retval_failed -eq 0 ]; then
    echo "Job failed. Please check logs."
    exit 1
fi
