#!/usr/bin/env bash

NAMESPACE=test
TIMEOUT_IN_SECONDS=120

echo "Waiting for tests to finish"
elapsedSeconds=0

status=$(kubectl get pod/application-gateway-test -n $NAMESPACE -ojsonpath='{ .status.phase}')
until [[ $status == Succeeded || $status == Failed ]]; do
  printf '.'
  sleep 5
  let "elapsedSeconds=elapsedSeconds+5"
  if [[ $elapsedSeconds -gt $TIMEOUT_IN_SECONDS ]] ; then
    exit 0
  fi
  status=$(kubectl get pod/application-gateway-test -n $NAMESPACE -ojsonpath='{ .status.phase}')
  echo $status
done
