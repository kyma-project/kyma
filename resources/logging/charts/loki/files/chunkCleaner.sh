#!/usr/bin/env bash
set -e

function updateCurrentSize() {
    CURRENT_SIZE_KB=$(df /data | tail -n 1 | awk '{print $3}')
}

trap 'EXIT=true' SIGTERM SIGINT

echo "VOLUME_SIZE: ${VOLUME_SIZE}"
VOLUME_SIZE_GB=$(echo "${VOLUME_SIZE}" | sed 's|Gi||g')
echo "VOLUME_SIZE_GB: ${VOLUME_SIZE_GB} GB"
VOLUME_SIZE_KB=$(( ${VOLUME_SIZE_GB} * 1024 * 1024 ))
echo "VOLUME_SIZE_KB: ${VOLUME_SIZE_KB} KB"

echo "TARGET_SIZE_PCT: ${TARGET_SIZE_PCT} %"
TARGET_SIZE_KB=$(( ${VOLUME_SIZE_KB} - (${VOLUME_SIZE_KB} * ${TARGET_SIZE_PCT} / 100) ))
echo "TARGET_SIZE_KB: ${TARGET_SIZE_KB} KB"

updateCurrentSize
echo "CURRENT_SIZE_KB: ${CURRENT_SIZE_KB} KB"

while [ -z "${EXIT}" ]
do
    updateCurrentSize
    while [ "${TARGET_SIZE_KB}" -lt "${CURRENT_SIZE_KB}" ] && [ -z "${EXIT}" ]
    do
        OLDEST=$(ls -t /data/loki/chunks | tail -1)
        echo "Current size ${CURRENT_SIZE_KB}KB is bigger then ${TARGET_SIZE_KB}KB, deleting file /data/loki/chunks/${OLDEST}"
        rm "/data/loki/chunks/${OLDEST}"
        updateCurrentSize
    done
    sleep "${SLEEP}"
done

echo "Application stopped"
