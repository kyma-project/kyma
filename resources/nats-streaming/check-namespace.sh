#!/usr/bin/env bash
counter=0
until [ $counter -gt 3 ]; do  # it tries 3 times to solve the "Terminating" status
  timestamp=`date`
  if errormessage=`kubectl create ns knative-eventing 2>&1`; then
    echo "[$timestamp] OK: knative-eventing created"
    exit 0
  elif echo "$errormessage" | grep 'already exists'; then
    echo "[$timestamp] OK: knative-eventing already exists"
    exit 0
  else
    echo "[$timestamp] ERROR: knative-eventing creation failed"
  fi
  sleep 5
  ((counter++))
done
exit 1
