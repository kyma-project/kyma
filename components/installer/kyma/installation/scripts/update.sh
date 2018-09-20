#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOCAL=0

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --local)
            LOCAL=1
            shift
            ;;
        --cr-name)
            CR_NAME="$2"
            shift # past argument
            shift # past value
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [ -z "$CR_NAME" ]; then
    echo "[ERR] Please provide installation custom resource name via --cr-name parameter"
    exit 1
fi

if [ $LOCAL -eq 1 ]; then
    bash $CURRENT_DIR/copy-resources.sh
fi

kubectl label installation/${CR_NAME} action=update --overwrite