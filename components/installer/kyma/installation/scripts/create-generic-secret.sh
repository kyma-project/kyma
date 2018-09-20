#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

function fillTemplate {
  local OUTPUT="$1"
  local KEY="$2"
  local VALUE="$3"
  local VALUE_BASE64=$(echo -n "${VALUE}" | base64 | tr -d '\n')

  case `uname -s` in
      Darwin)
          sed -i "" "s;__${KEY}__;${VALUE_BASE64};" "${OUTPUT}"
          ;;
      *)
          sed -i "s;__${KEY}__;${VALUE_BASE64};g" "${OUTPUT}"
          ;;
  esac
}

OUTPUT="$1"
shift

TPL_PATH="${OUTPUT}.tpl"
cp $TPL_PATH $OUTPUT

while [[ $# -gt 0 ]]
do
    KEY="$1"
    VALUE="$2"

    fillTemplate  "${OUTPUT}" "${KEY}" "${VALUE}"
    shift
    shift
done
