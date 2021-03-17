#!/usr/bin/env bash
# Script for build preview of this repo like in https://kyma-project.io/docs/ on every PR.
# For more information, please contact with: @m00g3n @aerfio @pPrecel @magicmatatjahu

set -eo pipefail

on_error() {
    echo -e "${RED}✗ Failed${NC}"
    exit 1
}
trap on_error ERR

readonly KYMA_PROJECT_IO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

readonly WEBSITE_DIR="website"
readonly WEBSITE_REPO="https://github.com/dbadura/website"

readonly BUILD_DIR="${KYMA_PROJECT_IO_DIR}/${WEBSITE_DIR}"
readonly PUBLIC_DIR="${KYMA_PROJECT_IO_DIR}/${WEBSITE_DIR}/public"
readonly DOCS_DIR="$( cd "${KYMA_PROJECT_IO_DIR}/../docs" && pwd )"

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

remove-cached-content() {
  ( rm -rf "${BUILD_DIR}" ) || true
}

copy-website-repo() {
  git clone -b "docs" --single-branch "${WEBSITE_REPO}" "${WEBSITE_DIR}"
}

build-preview() {
  export APP_PREVIEW_SOURCE_DIR="${KYMA_PROJECT_IO_DIR}/.."
  export APP_DOCS_BRANCHES="preview"
  export APP_PREPARE_FOR_REPO="kyma"
  make -C "${BUILD_DIR}" netlify-docs-preview
}

copy-build-result() {
  mkdir -p "${DOCS_DIR}/.kyma-project-io/"
  cp -rp "${PUBLIC_DIR}/" "${DOCS_DIR}/.kyma-project-io/"
}

main() {
  step "Remove website cached content"
  remove-cached-content
  pass "Removed"

  step "Copying kyma/website repo"
  copy-website-repo
  pass "Copied"

  step "Building preview"
  build-preview
  pass "Builded"

  step "Copying public folder"
  copy-build-result
  pass "Copied"
}
main
