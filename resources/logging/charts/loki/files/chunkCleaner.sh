#!/usr/bin/env bash
set -e

function updateCurrentSize() {
    CURRENT_SIZE_KB=$(df /data | tail -n 1 | awk '{print $3}')
}

trap 'EXIT=true' SIGTERM SIGINT

AVAILABLE_SIZE_KB=$(df /data | tail -n 1 | awk '{print $4}')
echo "AVAILABLE_SIZE_KB: ${AVAILABLE_SIZE_KB} KB"

updateCurrentSize
echo "CURRENT_SIZE_KB: ${CURRENT_SIZE_KB} KB"

VOLUME_SIZE_KB=$(( ${AVAILABLE_SIZE_KB} + ${CURRENT_SIZE_KB} ))
echo "VOLUME_SIZE_KB: ${VOLUME_SIZE_KB} KB"

echo "TARGET_SIZE_PCT: ${TARGET_SIZE_PCT} %"
TARGET_SIZE_KB=$(( ${VOLUME_SIZE_KB} - (${VOLUME_SIZE_KB} * ${TARGET_SIZE_PCT} / 100) ))
echo "TARGET_SIZE_KB: ${TARGET_SIZE_KB} KB"

while [ -z "${EXIT}" ]
do
    updateCurrentSize
    while [ "${TARGET_SIZE_KB}" -lt "${CURRENT_SIZE_KB}" ] && [ -z "${EXIT}" ]
    do
        echo "Current size ${CURRENT_SIZE_KB}KB is bigger then ${TARGET_SIZE_KB}KB, deleting ${BATCH_SIZE} files"
        for OLDEST in $(ls -t /data/loki/chunks | tail -${BATCH_SIZE})
        do
            echo "Deleting file /data/loki/chunks/${OLDEST}"
            rm "/data/loki/chunks/${OLDEST}"
        done
        updateCurrentSize
    done
    sleep "${SLEEP}"
done

echo "Application stopped"
