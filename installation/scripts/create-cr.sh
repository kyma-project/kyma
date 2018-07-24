#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CRTPL_PATH="$CURRENT_DIR/../resources/installer-cr.yaml.tpl"

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --url)
            URL="$2"
            shift # past argument
            shift # past value
            ;;
        --output)
            OUTPUT="$2"
            shift
            shift
            ;;
        --version)
            VERSION="$2"
            shift
            shift
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

cp $CRTPL_PATH $OUTPUT

case `uname -s` in
    Darwin)
        sed -i "" "s/__VERSION__/${VERSION}/" "$OUTPUT"
        sed -i "" "s;__URL__;${URL};" "$OUTPUT"
        ;;
    *)
        sed -i "s/__VERSION__/${VERSION}/g" "$OUTPUT"
        sed -i "s;__URL__;${URL};g" "$OUTPUT"
        ;;
esac
