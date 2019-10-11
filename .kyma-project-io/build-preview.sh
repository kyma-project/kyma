#!/usr/bin/env bash
# Script for preparing content for kyma-project.io

set -e
set -o pipefail

pushd "$(pwd)" > /dev/null

on_error() {
    echo -e "${RED}✗ Failed${NC}"
    exit 1
}
trap on_error ERR

on_exit() {
    popd > /dev/null
}
trap on_exit EXIT

readonly KYMA_PROJECT_IO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly WEBSITE_DIR="website"
readonly WEBSITE_REPO="https://github.com/magicmatatjahu/website"

readonly BUILD_DIR="${KYMA_PROJECT_IO_DIR}/${WEBSITE_DIR}"

#"https://github.com/kyma-project/website"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[0;33m'
readonly NC='\033[0m' # No Color

pass() {
    local message="$1"
    echo -e "${GREEN}√ ${message}${NC}"
}

step() {
    local message="$1"
    echo -e "\\n${YELLOW}${message}${NC}"
}

copy-website-repo() {
  git clone -b "docs-community-preview" --single-branch "${WEBSITE_REPO}" "${WEBSITE_DIR}"
}

pre-build() {
  cd "${BUILD_DIR}" && make resolve
  cd "${BUILD_DIR}" && make clear-cache
  cd "${BUILD_DIR}" && make prepare-tools
}

prepare-content() {
    export APP_PREPARE_DOCS="true"
    export APP_PREPARE_COMMUNITY="false"
    export APP_PREPARE_ROADMAP="false"
    export APP_DOCS_SOURCE_DIR="${KYMA_PROJECT_IO_DIR}/.."
    export APP_DOCS_OUTPUT="${BUILD_DIR}/content/docs/netlify-preview"

    cd "${BUILD_DIR}" && make -C "./tools/content-loader" fetch-content
}

build-preview() {
  cd "${BUILD_DIR}" && make build
}

main() {
  step "Copying kyma/website repo"
  copy-website-repo
  pass "Copied"

  step "Pre building process"
  pre-build
  pass "Processed"

  step "Preparing content"
  prepare-content
  pass "Prepared"

  step "Building preview"
  build-preview
  pass "Builded"
}
main