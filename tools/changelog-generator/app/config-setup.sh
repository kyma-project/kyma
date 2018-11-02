#!/usr/bin/env sh

echo "Copying configuration file..."
cp ${CONFIG_FILE} ./package.json

function removingLatestTag() {
    LATEST_TAG_EXISTS=false
    if [ $(git tag -l "$LATEST_TAG") ]; then
        LATEST_TAG_EXISTS=true
        echo "Temporary removing 'latest' tag..."
        LATEST_TAG_REV=$(git rev-list -n 1 ${LATEST_TAG})
        LATEST_TAG_MESSAGE=$(git tag -l --format='%(contents)' ${LATEST_TAG})
        git tag -d ${LATEST_TAG}
    fi
}

if [ "$SKIP_REMOVING_LATEST" != "true" ]; then
    removingLatestTag
fi
