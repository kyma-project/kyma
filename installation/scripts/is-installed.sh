#!/bin/bash

#################################################################################
#
# Follow the process of Kyma installation
# Returns exit code 0 if Kyma is installed successfully
#
# Options:
# --verbose - reduce the delay between status checks; print the installation CR
#             and status of all pods in every check 
# --timeout - set the timeout after which the logs from installer will be printed
#             and script will return exit code 1; the time accuracy depends on
#             the delay between status checks; after the flag provide the time
#             in the seconds (s/S), minutes (m/M) or hours (h/H) in the following
#             format: <number_of_time_units><time_unit>
#             
# Samples: 
#             bash is-installed.sh --timeout 25m
#             bash is-installed.sh --verbose
#
#################################################################################

function parse_timeout {
  TIME=$(echo "$TIMEOUT" | tr -dc '0-9')
  UNIT=$(echo "$TIMEOUT" | tr -d  '0-9' | tr -d ' ')

  case "${UNIT}" in
    "s"|"S") TIMEOUT=TIME ;;
    "m"|"M") TIMEOUT=$(( TIME*60 )) ;;
    "h"|"H") TIMEOUT=$(( TIME*3600 )) ;;
    "")
      echo "time unit not provided"
      exit 1
      ;;
    *)
      echo "unknown time unit: ${UNIT}"
      exit 1
      ;;
  esac
}

TIMEOUT=""
TIMEOUT_SET=0
ITERATIONS_LEFT=0
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
        --timeout)
          TIMEOUT_SET=1
          TIMEOUT="$2"
          shift # past flag
          shift # past value
          ;;
        *)    # unknown option
          POSITIONAL+=("$1") # save it in an array for later
          shift # past argument
          ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [ "$TIMEOUT_SET" -ne 0 ]; then
  parse_timeout
  ITERATIONS_LEFT=$(( TIMEOUT/DELAY ))
fi

echo "Checking state of kyma installation...hold on"
while :
do
  STATUS="$(kubectl -n default get installation/kyma-installation -o jsonpath='{.status.state}')"
  DESC="$(kubectl -n default get installation/kyma-installation -o jsonpath='{.status.description}')"
  if [ "$STATUS" = "Installed" ]
  then
      echo "kyma is installed..."
      break
  elif [ "$STATUS" = "Error" ]
  then
    echo "kyma installation error, description: ${DESC}"
    if [ "$TIMEOUT_SET" -eq 0 ]; then
      echo "to fetch the logs from the installer execute: kubectl logs -n kyma-installer $(kubectl get pods --all-namespaces -l name=kyma-installer --no-headers -o jsonpath='{.items[*].metadata.name}')"
    fi
    if [ "$TIMEOUT_SET" -ne 0 ] && [ "$ITERATIONS_LEFT" -le 0 ]; then
      echo "Installation errors until timeout:"
      echo "----------"
      kubectl -n default get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}
{{.component}}:
  {{.log}}
{{- end}}
'
      echo "----------"
      echo "timeout reached on kyma installation error. Exiting"
      exit 1
    fi
  else 
    echo "Status: ${STATUS}, description: ${DESC}"
    if [ "$TIMEOUT_SET" -ne 0 ] && [ "$ITERATIONS_LEFT" -le 0 ]; then
      echo "----------"
      kubectl -n default get installation kyma-installation -o go-template --template='{{- range .status.errorLog }}
{{.component}}:
  {{.log}}
{{- end}}
'
      echo "----------"
      echo "timeout reached. Exiting"
      exit 1
    fi
  fi
  if [ "${VERBOSE}" -eq 1 ]; then
    kubectl -n default get installation/kyma-installation -o yaml
    echo "----------"
    kubectl get po --all-namespaces
  fi
  sleep $DELAY
  ITERATIONS_LEFT=$(( ITERATIONS_LEFT-1 ))
done
