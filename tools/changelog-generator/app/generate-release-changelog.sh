#!/usr/bin/env sh

source ${APP_PATH}/variables.sh

CONFIGURE_GIT=$1
if [ "$CONFIGURE_GIT" = "--configure-git" ]; then
    source ${APP_PATH}/config-git.sh
fi

source ${APP_PATH}/config-setup.sh

mkdir -p ${RELEASE_CHANGELOG_FILE_DIRECTORY}
if [ -n "$FROM_TAG" ]; then
    echo "Getting new changes starting from the '${FROM_TAG}' tag to '${LATEST_VERSION}' tag..."
    lerna-changelog --from=${FROM_TAG} --to=${LATEST_VERSION} | sed -e "s/## Unreleased/## ${LATEST_VERSION}/g" > ${RELEASE_CHANGELOG_FILE_PATH}
else
    REVISION=$(git rev-list --tags --max-count=1)
    LAST_VERSION_TAG=$(git describe --tags ${REVISION});
    echo "Getting new changes starting from the '${LAST_VERSION_TAG}' tag..."
    lerna-changelog --from=${LAST_VERSION_TAG} | sed -e "s/## Unreleased/## ${LATEST_VERSION}/g" > ${RELEASE_CHANGELOG_FILE_PATH}
fi

source ${APP_PATH}/config-cleanup.sh
