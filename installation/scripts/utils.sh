#!/usr/bin/env bash

#
# Log is function useful for logs creation.
# It accepts three arguments:
# $1 - text we want to print
# $2 - text color
# $3 - text style
# It will create single log with defined color and font style.
# To specify style without color we have to put 'nc' before style.
# For example:
# log "gophers" magenta bold - will print bold 'gophers' in magenta color.
# log "text" bold - it will print normal text.
# log "text" nc bold - it will print bold text.
# By default log will print normal text like echo command.
# Use source [utils.sh path] to import log function into your script.
#

function log() {
    local exp=$1;
    local color=$2;
    local style=$3;
    local NC='\033[0m'
    if ! [[ ${color} =~ '^[0-9]$' ]] ; then
       case $(echo ${color} | tr '[:upper:]' '[:lower:]') in
        black) color='\e[30m' ;;
        red) color='\e[31m' ;;
        green) color='\e[32m' ;;
        yellow) color='\e[33m' ;;
        blue) color='\e[34m' ;;
        magenta) color='\e[35m' ;;
        cyan) color='\e[36m' ;;
        white) color='\e[37m' ;;
        nc|*) color=${NC} ;; # no color or invalid color
       esac
    fi
    if ! [[ ${style} =~ '^[0-9]$' ]] ; then
        case $(echo ${style} | tr '[:upper:]' '[:lower:]') in
        bold) style='\e[1m' ;;
        underline) style='\e[4m' ;;
        inverted) style='\e[7m' ;;
        *) style="" ;; # no style or invalid style
       esac
    fi
    printf "${color}${style}${exp}${NC}\n"
}

#
# showFailedResources is function to print some logs and pods details
# It accepts on argument:
# $1 - namespace from which we want to get details
# It will 
# - display list of pods and PVCs, 
# - describe pods with status != running
# - display logs for not running pods
# - describe PVCs with status != Bound
#

function showFailedResources {
  local ns=$1
  kubectl get pods,pvc -n $1 -o wide
  
  notRunningPods=($(kubectl get pods -n $1 -o=custom-columns=NAME:.metadata.name,STATUS:.status.phase  | grep -v Running | awk '{if(NR>1)print $1}'))
  if [[ -n ${notRunningPods-} ]];
  then
    for i in "${notRunningPods[@]}"
    do
      echo "======================================================================================"
      echo "kubectl describe pod $i"
      echo "======================================================================================"
      kubectl describe pod $i -n $1
      containers=($(kubectl get pods -n $1 $i  -o jsonpath='{range .status.containerStatuses[*]}{.name}{"\t"}{.ready}{"\n"}' | grep "\ttrue" |  awk '{if(NR>1)print $1}'))
      if [[ -n ${containers-} ]];
      then
      for j in "${containers[@]}"
        do
          echo "=========="
          echo "kubectl logs $i -c $j"
          echo "=========="
          kubectl logs -n $1 --tail=100 $i -c $j
        done
      fi
    done
  fi

  notBoundPvc=($(kubectl get pvc -n $1 -o=custom-columns=NAME:.metadata.name,STATUS:.status.phase  | grep -v Bound | awk '{if(NR>1)print $1}'))
  if [[ -n ${notBoundPvc-} ]];
  then
    for i in "${notBoundPvc[@]}"
    do
      echo "======================================================================================"
      echo "kubectl describe pvc $i"
      echo "======================================================================================"
      kubectl describe pvc $i -n $1
    done
  fi
}

# checkInputParameterValue is a function to check if input parameter is valid
# There HAS to be provided argument:
# $1 - value for input parameter
# for example in installation/cmd/run.sh we can set --vm-driver argument, which has to have a value.

function checkInputParameterValue() {
    if [ -z "${1}" ] || [ "${1:0:2}" == "--" ]; then
        echo "Wrong parameter value"
        echo "Make sure parameter value is neither empty nor start with two hyphens"
        exit 1
    fi
}