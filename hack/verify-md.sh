#!/usr/bin/env bash

echo "${BASH_SOURCE[0]}"

readonly CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/hack"
readonly ROOT_PATH="$( cd "${CURRENT_DIR}/../" && pwd )"

TMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "${TMP_DIR}" || true
}

trap cleanup EXIT SIGINT

cp -a ${ROOT_PATH}/docs/05-technical-reference/00-custom-resources/. $TMP_DIR/

cd $CURRENT_DIR/table-gen || exit

go run $CURRENT_DIR/table-gen/table-gen.go --crd-filename ../../installation/resources/crds/telemetry/tracepipelines.crd.yaml --md-filename $TMP_DIR/telemetry-03-tracepipeline.md --crd-title TracePipeline

go run $CURRENT_DIR/table-gen/table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename $TMP_DIR/telemetry-01-logpipeline.md --crd-title LogPipeline

go run $CURRENT_DIR/table-gen/table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logparsers.crd.yaml --md-filename $TMP_DIR/telemetry-02-logparser.md --crd-title LogParser

DIFF=$(diff -q $TMP_DIR ${ROOT_PATH}/docs/05-technical-reference/00-custom-resources)
if [ -n "${DIFF}" ]; then 
    echo -e "ERROR: there is a difference between operator CRD and documentation"
    echo -e "Please, go to the hack/table-gen, and run 'make run'"
    exit 1
fi