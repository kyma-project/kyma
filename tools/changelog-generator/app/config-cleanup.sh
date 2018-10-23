#!/usr/bin/env sh

echo "Cleaning up..."
rm ./package.json

if [ "$LATEST_TAG_EXISTS" = "true" ]; then
    echo "Restoring 'latest' tag for revision ${LATEST_TAG_REV}..."
    git tag -a ${LATEST_TAG} ${LATEST_TAG_REV} -m "${LATEST_TAG_MESSAGE}"
fi
