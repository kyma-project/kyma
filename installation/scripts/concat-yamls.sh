#!/usr/bin/env bash

set -o errexit

for file in "$@"
do

    TMP=""

    if [[ ! -f "${file}" ]]; then
        echo "File ${file} not found"
        exit 1
    fi

    TMP=$(cat "${file}" | sed '/^\s*$/d')

    FIRST_LINE=$(head -n1 <<< "${TMP}")
    if [[ "$FIRST_LINE" == "---" ]]; then
        TMP=$(sed '1d' <<< "${TMP}")
    fi

    echo "${TMP}"

    LAST_LINE=$(tail -n1 <<< "${TMP}")
    if [[ "$LAST_LINE" != "---" ]]; then
        echo '---'
    fi

done