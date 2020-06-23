#!/usr/bin/env bash
seconds=48    #wait 4mins till "knative-eventing" namespace will be created by another Helm3 installer. Helm3 default timeout 5mins
counter=0
until [ $counter -gt $seconds ]; do
  timestamp=`date`
  if kubectl get ns | grep knative-eventing; then
    echo "[$timestamp] knative-eventing found"
    break
  else
    echo "[$timestamp] knative-eventing not found"
  fi
  sleep 5
  ((counter++))
done
