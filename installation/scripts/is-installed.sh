#!/bin/bash

VERBOSE=0
DELAY=10

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --verbose)
          VERBOSE=1
          DELAY=5
          shift # past flag
          ;;
        *)    # unknown option
          POSITIONAL+=("$1") # save it in an array for later
          shift # past argument
          ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo "Checking state of kyma installation...hold on"
while :
do
  STATUS="$(kubectl get installation/kyma-installation -o jsonpath='{.status.state}')"
  DESC="$(kubectl get installation/kyma-installation -o jsonpath='{.status.description}')"
  if [ "$STATUS" = "Installed" ]
  then
      echo "kyma is installed..."
      break
  elif [ "$STATUS" = "Error" ]
  then
    echo "kyma installation error... ${DESC}"
    echo "----------"
    echo "$(kubectl logs -n kyma-installer $(kubectl get pods --all-namespaces -l name=kyma-installer --no-headers -o jsonpath='{.items[*].metadata.name}'))"
    echo "----------"
    exit 1
  else 
    echo "Status: ${STATUS}, description: ${DESC}"
    if [ "${VERBOSE}" -eq 1 ]; then
      echo "$(kubectl get installation/kyma-installation -o yaml)"
      echo "----------"
      echo "$(kubectl get po --all-namespaces)"
    fi
    sleep $DELAY
  fi
done