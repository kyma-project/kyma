#!/usr/bin/env bash
#
# Script for generating and publishing new documentation to kyma-project.io

pushd "$(pwd)" > /dev/null || exit

on_error() {
    echo -e "${RED}✗ Failed${NC}"
    exit 1
}
trap on_error ERR

on_exit() {
    popd > /dev/null || exit
}
trap on_exit EXIT

readonly SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
readonly KYMA_REPOSITORY="https://github.com/kyma-project/kyma.git"
readonly DOCUMENTATION_DIR="${SCRIPTS_DIR}/../static/documentation"
readonly GENERATOR_IMAGE="eu.gcr.io/kyma-project/documentation-generator:0.1.49"

# Colors
readonly ARGS=("$@")
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

init() {
    PUBLISH='false'

    while test $# -gt 0; do
        case "$1" in
            --publish | -p)
                PUBLISH='true'
                shift
                ;;
            --skip | -sk)
                shift
                if [[ ${1} == true ]]; then
                    echo 'Aborting the build to prevent a loop'
                    exit 0
                fi
                shift
                ;;
            --help | -h)
                exit 0
                ;;
            *)
                echo "$1 is not a recognized flag!"
                exit 1
                ;;
        esac
    done
    readonly PUBLISH
}

generate() {
    docker run --rm -v "${DOCUMENTATION_DIR}:/app/documentation" \
               -e APP_OUTPUT="/app/documentation" \
               -e APP_DOC_CONFIG_FILE="/app/documentation/config.json" \
               -e APP_TOKEN="${BOT_GITHUB_TOKEN}" \
               ${GENERATOR_IMAGE} || return
}

publish() {
    echo "Detecting changes"
    local changes
    changes=$(git status --porcelain "${DOCUMENTATION_DIR}" | wc -w)
    if [[ "${changes}" -eq 0 ]]; then
        echo "Nothing to publish"
        return
    fi

    echo "Commit documentation"
    git add "${DOCUMENTATION_DIR}" || return
    # git commit -m "Publish documentation for Kyma" --no-verify || return

    echo "Pushing documentation to master"
    # git push origin HEAD:master || return
}

main() {
    init "${ARGS[@]}" || return

    cd /home/prow/go/src/github.com/kyma-project/website
echo $(pwd)
echo $(ls)
echo "HOME ${HOME}"
    step "Generating"
    generate || return
    pass "Generated"

    # step "Publishing"
    # if [[ ${PUBLISH} == true ]]; then
    #     publish || return
    # else
    #     echo "Skipped"
    # fi
    # pass "Published"
}
main