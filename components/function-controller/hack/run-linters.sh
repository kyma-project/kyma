#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BIN_DIR="${SCRIPT_DIR}/../bin"
readonly GOLANGCI_LINT_BINARY="${BIN_DIR}/golangci-lint"

readonly GOLANGCI_LINT_VERSION="v1.23.1"

main(){
    if [[ ! -x "${GOLANGCI_LINT_BINARY}" ]]; then
        echo "Downloading golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${BIN_DIR}" "${GOLANGCI_LINT_VERSION}"
        echo "Done!"
    fi
    
    "${GOLANGCI_LINT_BINARY}" run --config "${SCRIPT_DIR}/../.golangci.yml"
}

main