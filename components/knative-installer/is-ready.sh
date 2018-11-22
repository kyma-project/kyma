#!/bin/bash

################################################################################
#
# Validate if specified POD is up and ready
# $1 - namespace
# $2 - pod's label name
# $3 - pod's label value
# Sample: bash isready.sh kube-system tiller
#
################################################################################

#Checking if POD is already deployed
trap "exit" INT
while :
do
  if [[ $(kubectl get pods -n "$1" -l "$2"="$3" -o jsonpath='{.items[*].metadata.name}') ]]
    then
      echo "$3 is deployed..."
      break
    else
      echo "$3 is not deployed - waiting 5s..."
      sleep 5
    fi
done

QUERY_PODS=true

while [[ "$QUERY_PODS" == true ]]
do
  #Checking if POD is ready to operate
  for POD in $(kubectl get pods -n "$1" -l "$2"="$3" -o jsonpath='{.items[*].metadata.name}')
  do
    QUERY_PODS=false
    trap "exit" INT
    while :
    do
      STATUS=$(kubectl get pod "$POD" -n "$1" -o jsonpath='{.status.containerStatuses[0].ready}' 2>&1)
      if [ "$STATUS" = "true" ]
      then
        echo "$POD is running..."
        break
      elif [[ $STATUS == *NotFound* ]]
      then
        # The pod probably no longer exists and we need to query again
        echo "$POD no longer exists..."
        QUERY_PODS=true
        break 2
      else
        echo "$POD is not running -  waiting 5s..." $(kubectl get event -n "$1" -o go-template='{{range .items}}{{if eq .involvedObject.name "'$POD'"}}{{.message}}{{"\n"}}{{end}}{{end}}' | tail -1)
        sleep 5
      fi
    done
  done
done

#checking only if kube-dns is checked
if [ "$3" = "kube-dns" ]
then

  for POD in $3
  do
    trap "exit" INT
    while :
    do
      if [[ "$(kubectl get ep $3 -n $1 -o jsonpath='{.subsets[0].addresses[0].ip}')" ]]
      then
        echo "kubedns endpoint IP assigned"
        break
      else
        echo "kubedns endpoint IP is not assigned yet -  waiting 5s..."
        sleep 5
      fi
    done
  done

fi
