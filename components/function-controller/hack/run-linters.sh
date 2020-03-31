#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly BIN_DIR="${SCRIPT_DIR}/../bin"
readonly GOLANGCI_LINT_BINARY="${BIN_DIR}/golangci-lint"

readonly GOLANGCI_LINT_VERSION="1.24.0"

main(){
    if [[ ! -x "${GOLANGCI_LINT_BINARY}" || "$(${GOLANGCI_LINT_BINARY} --version | grep "${GOLANGCI_LINT_VERSION}" -c)"==1 ]]; then
        echo "Downloading golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${BIN_DIR}" "v${GOLANGCI_LINT_VERSION}"
        echo "Done!"
    fi
    "${GOLANGCI_LINT_BINARY}" run --config "${SCRIPT_DIR}/../.golangci.yml"
}

main