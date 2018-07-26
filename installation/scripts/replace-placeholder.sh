#!/bin/bash

set -o errexit

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --path)
            FILE_PATH="$2"
            shift
            ;;
        --placeholder)
            PLACEHOLDER="$2"
            shift
            shift
            ;;
        --value)
            VALUE="$2"
            shift
            shift
            ;;
        *) # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

if [[ -z $FILE_PATH ]] ; then
    (>&2 echo "Path not provided")
    exit 1
fi

if [[ -z $PLACEHOLDER ]] ; then
    (>&2 echo "Placeholder not provided")
    exit 1
fi

case `uname -s` in
    Darwin)
        sed -i "" "s;${PLACEHOLDER};${VALUE};" "${FILE_PATH}"
        ;;
    *)
        sed -i "s;${PLACEHOLDER};${VALUE};g" "${FILE_PATH}"
        ;;
esac