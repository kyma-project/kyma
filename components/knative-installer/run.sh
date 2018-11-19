#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

case "$ACTION" in
    "install")
        ${DIR}/install.sh
        ;;
    "configure")
        ${DIR}/configure.sh
        ;;
esac