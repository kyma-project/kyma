#!/bin/bash

OPERATION=""
DELIMITER=""
ORIGINAL_TEXT=""

while [[ $# -gt 0 ]] ; do
    case "$1" in 
        --prefix)
            OPERATION="PREFIX"
            shift
            ;;
        --suffix)
            OPERATION="SUFFIX"
            shift
            ;;
        --delimiter | -d)
            DELIMITER=$2
            shift
            shift
            ;;
        *)
            break
            ;;
    esac
done

if [[ -z $DELIMITER ]] ; then
    (>&2 echo "Delimiter not provided")
    exit 1
fi

ORIGINAL_TEXT=$1
if [[ -z $ORIGINAL_TEXT ]]; then
    (>&2 echo "Text to cut not provided")
    exit 1
fi

if [[ "$OPERATION" = "PREFIX" ]]; then
    echo "${ORIGINAL_TEXT#*$DELIMITER}" # removes prefix ending with delimiter
elif [[ "$OPERATION" = "SUFFIX" ]]; then 
    echo "${ORIGINAL_TEXT%$DELIMITER*}" # removes suffix starting with delimiter
else
    (>&2 echo "Operation not specified. Please provide flag \"--prefix\" or \"--suffix\"")
    exit 1
fi