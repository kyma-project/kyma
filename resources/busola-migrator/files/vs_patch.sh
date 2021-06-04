#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# sed command used to append suffix to Virtual Service host
SED_COMMAND="s/^[^\.]+/&${SUFFIX}/"

# patch_vs patches first Virtual Service host, adding $SUFFIX to the subdomain
# $1 - Virtual Service name
# $2 - Virtual Service namespace
patch_vs() {
  # get host name from VS
  OLD_HOST=$(kubectl get virtualservice -n "$2" "$1" -o=jsonpath="{.spec.hosts[0]}" 2>/tmp/errmsg) && EXIT_CODE=$? || EXIT_CODE=$?
  ERR_MSG=$(cat /tmp/errmsg)
  # if couldn't get VS
  if [[ $EXIT_CODE -ne "0" ]]; then
    if [[ $ERR_MSG == *"Error from server (NotFound)"* ]]; then
      # VS not found
      echo "VirtualService [$2/$1] not found"
      return 0
    else
      # other errors while getting VS
      echo "$ERR_MSG"
      exit $EXIT_CODE
    fi
  fi

  # check if VS wasn't already patched
  SUB=$(echo -n "$OLD_HOST" | grep -Eoh '^[^\.]+')
  if [[ $SUB == *"$SUFFIX" ]]; then
    echo "VirtualService [$2/$1] already patched"
    return 0
  fi

  # patch VS
  NEW_HOST=$(echo -n "$OLD_HOST" | sed -E "$SED_COMMAND")
  kubectl patch virtualservice -n "$2" "$1" --type merge --patch '{"spec": {"hosts": ["'"${NEW_HOST}"'"]}}'
}

patch_vs "$VS_DEX_NAME" "$VS_DEX_NAMESPACE"
patch_vs "$VS_CONSOLE_NAME" "$VS_CONSOLE_NAMESPACE"
