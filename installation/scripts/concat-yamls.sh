#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
COMBO_FILE=$(mktemp ${CURRENT_DIR}/../../temp.XXXXXX)

for file in "$@"
do
    if [[ ! -f ${file} ]]; then
        echo "File ${file} is not a file. Terminating script."
        rm "${COMBO_FILE}"
        exit 1
    fi

    FIRST_LINE_OF_YAML=$(head -n 1 "${file}")
    if [[ $FIRST_LINE_OF_YAML == "---" ]]; then
        cat ${file} | sed '1d' >> ${COMBO_FILE}
    else
        cat ${file} >> ${COMBO_FILE}
    fi

    LAST_LINE_OF_COMBO=$(tail -n 1 "${COMBO_FILE}")
    if [[ $LAST_LINE_OF_COMBO != "---" ]]; then
        printf '\n---' >> ${COMBO_FILE}
    fi
    printf '\n' >> ${COMBO_FILE}
    
done

case `uname -s` in
    Darwin)
        sed -i '' '/^\s*$/d' ${COMBO_FILE}
        ;;
    *)
        sed -i '/^\s*$/d' ${COMBO_FILE}
        ;;
esac

cat "${COMBO_FILE}"
rm "${COMBO_FILE}"