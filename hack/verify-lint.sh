#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

readonly CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
readonly ROOT_PATH="${1:-$( cd "${CURRENT_DIR}/.." && pwd )}" # first argument or root of the project
readonly TMP_DIR=$(mktemp -d)
readonly GOLANGCI_LINT_VERSION="v1.38.0"

source "${CURRENT_DIR}/utilities.sh" || { echo 'Cannot load CI utilities.'; exit 1; }

cleanup() {
    rm -rf "${TMP_DIR}" || true
}

trap cleanup EXIT SIGINT

# verify_installation checks if golangci-lint is installed and if its version is at least the expected one.
# If the check does not pass, installation instructions are printed.
golangci::verify_installation() {
  # if binary found check version
  if [ ! -z "$(command -v golangci-lint)" ]; then
    local CURRENT_VERSION="$(golangci-lint version --format short 2>&1)"
    
    # remove the optional "v" prefix to only compare numbers
    EXPECTED_VERSION=${GOLANGCI_LINT_VERSION#"v"}
    CURRENT_VERSION=${CURRENT_VERSION#"v"}

    if [ "${EXPECTED_VERSION}" \> "${CURRENT_VERSION}" ]; then
      # Print instructions to update
      echo -e "${RED}x Installed golangci-lint version (${CURRENT_VERSION}) incorrect${NC}"
      echo -e "Please update to a version equal or greater than ${GOLANGCI_LINT_VERSION}"
      echo -e "Run the following command to update:"
      echo -e "${INVERTED}curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b <INSTALL_DIR> ${GOLANGCI_LINT_VERSION}${NC}"
      exit 1
    fi
    ## installed and version is correct
    return
  fi

  # not installed
  echo -e "${RED}x golangci-lint not installed${NC}"
  echo -e "Run the following command to install:"
  echo -e "${INVERTED}curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b <INSTALL_DIR> ${GOLANGCI_LINT_VERSION}${NC}"
}

golangci::run_checks() {
  shout "Run golangci-lint checks"
  LINTS=(
    # default golangci-lint lints
    deadcode errcheck gosimple govet ineffassign staticcheck \
    structcheck typecheck unused varcheck \
    # additional lints
    golint gofmt misspell gochecknoinits unparam scopelint gosec
  )

  echo "Checks: ${LINTS[*]}"
  cd $ROOT_PATH
  golangci-lint --disable-all --enable="$(sed 's/ /,/g' <<< "${LINTS[@]}")" --timeout=10m run --config $CURRENT_DIR/.golangci.yml

  echo -e "${GREEN}√ run golangci-lint${NC}"
}

main() {
  if [[ "${SKIP_VERIFY:-x}" != "true" ]]; then
    golangci::verify_installation
  fi

  golangci::run_checks
}

main
