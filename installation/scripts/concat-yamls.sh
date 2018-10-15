#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMBO_FILE=$(mktemp ${CURRENT_DIR}/../../temp.XXXXXX)
TMP_FILE=$(mktemp ${CURRENT_DIR}/../../temp.XXXXXX)

sedPattern () {

    case `uname -s` in
        Darwin)
            sed -i '' "$1" "$2"
            ;;
        *)
            sed -i "$1" "$2"
            ;;
    esac
}

for file in "$@"
do
    if [[ ! -f "${file}" ]]; then
        echo "File '${file}' not found"
        continue
    fi

    cat "${file}" > "${TMP_FILE}"
    sedPattern '/^\s*$/d' "${TMP_FILE}"

    FIRST_LINE=$(head -n 1 "${TMP_FILE}")
    if [[ "$FIRST_LINE" == "---" ]]; then
        sedPattern '1d' ${TMP_FILE}
    fi

    LAST_LINE=$(tail -n 1 "${TMP_FILE}")
    if [[ "$LAST_LINE" != "---" ]]; then
        echo '---' >> "${TMP_FILE}"
    fi

    cat "${TMP_FILE}" >> "${COMBO_FILE}"
    
done

cat "${COMBO_FILE}"
rm "${COMBO_FILE}"
rm "${TMP_FILE}"