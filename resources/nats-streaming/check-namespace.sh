#!/usr/bin/env bash
seconds=240    #seconds to wait till "knative-eventing namespace" will be created by another Helm 3 installer
counter=0
until [ $counter -gt $seconds ]; do
  timestamp=`date`
  if kubectl get ns | grep knative-eventing; then
    echo "[$timestamp] knative-eventing found"
    break
  else
    echo "[$timestamp] knative-eventing not found"
  fi
  sleep 1
  ((counter++))
done
