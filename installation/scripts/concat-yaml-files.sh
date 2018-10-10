#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PATH="${CURRENT_DIR}/../.."
TMPFILE=$(mktemp ${KYMA_PATH}/temp.XXXXXX)

for path in "$@"
do
    if [[ ! -f ${path} ]]; then
        echo "${path} is not a file. Terminating script."
        rm "${TMPFILE}"
        exit 1
    fi

    firstLine=$(head -n 1 "${path}")
    lastLine=$(tail -n 1 "${path}")

    if [[ "${firstLine}" != "---" ]] && [[ "${lastLine}" != "---" ]]; then
        cat "${path}" | sed '/^\s*$/d'>> ${TMPFILE}
    elif [[ "${firstLine}" == "---" ]] && [[ "${lastLine}" != "---" ]]; then
        sed '1d' "${path}" | sed '/^\s*$/d' >> ${TMPFILE}
    elif [[ "${firstLine}" != "---" ]] && [[ "${lastLine}" == "---" ]]; then
        sed '$d' "${path}" | sed '/^\s*$/d' >> ${TMPFILE}
    elif [[ "${firstLine}" == "---" ]] && [[ "${lastLine}" == "---" ]]; then
        sed '$d' "${path}" | sed '1d' | sed '/^\s*$/d' >> ${TMPFILE}
    fi

    echo '---' >> ${TMPFILE}
done

cat ${TMPFILE}
rm "${TMPFILE}"