#!/usr/bin/env sh

echo "Copying configuration file..."
cp ${APP_PATH}/package.json ./

LATEST_TAG_EXISTS=false
if [ $(git tag -l "$LATEST_TAG") ]; then
    LATEST_TAG_EXISTS=true
    echo "Temporary removing 'latest' tag..."
    LATEST_TAG_REV=$(git rev-list -n 1 ${LATEST_TAG})
    LATEST_TAG_MESSAGE=$(git tag -l --format='%(contents)' ${LATEST_TAG})
    git tag -d ${LATEST_TAG}
fi

