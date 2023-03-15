#!/usr/bin/env bash

readonly CURRENT_DIR="$( cd "$( dirname "$0" )" && pwd )"

trap cleanup EXIT SIGINT

cd $CURRENT_DIR/table-gen || exit

make generate

DIFF=$(git diff --exit-code ../docs/05-technical-reference)
if [ -n "${DIFF}" ]; then 
    echo -e "ERROR: there is a difference between operator CRD and documentation"
    echo -e "Please, go to the hack/table-gen, and run 'make run'"
    exit 1
fi