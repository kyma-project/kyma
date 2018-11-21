#!/usr/bin/env sh

source ${APP_PATH}/variables.sh

CONFIGURE_GIT=$1
if [ "$CONFIGURE_GIT" = "--configure-git" ]; then
    source ${APP_PATH}/config-git.sh
fi

source ${APP_PATH}/config-setup.sh

if [ -z "$FROM_TAG" ]; then
    REVISION=$(git rev-list --tags --max-count=1)
    FROM_TAG=$(git describe --tags ${REVISION});
fi

TO_ARGUMENT=""
if [ -n "$TO_TAG" ]; then
    TO_ARGUMENT="--to=${TO_TAG}"
    TO_ARGUMENT_MSG="to '${TO_TAG}' tag"
fi

mkdir -p ${RELEASE_CHANGELOG_FILE_DIRECTORY}
echo "Getting new changes starting from the '${FROM_TAG}' tag ${TO_ARGUMENT_MSG}..."
eval "lerna-changelog --from=${FROM_TAG} ${TO_ARGUMENT}" | sed -e "s/## Unreleased/## ${NEW_RELEASE_TITLE}/g" > ${RELEASE_CHANGELOG_FILE_PATH}

source ${APP_PATH}/config-cleanup.sh
