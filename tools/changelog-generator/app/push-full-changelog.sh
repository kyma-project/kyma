#!/usr/bin/env bash

source ${APP_PATH}/variables.sh

CONFIGURE_GIT=$1
if [ "$CONFIGURE_GIT" = "--configure-git" ]; then
    source ${APP_PATH}/config-git.sh
fi

git add ${FULL_CHANGELOG_FILE_PATH}
git commit -m "Update CHANGELOG.md for version ${NEW_RELEASE_TITLE}"

# Commit changelog also on master
if [ $BRANCH != "master" ]; then
    git fetch
    git checkout -f master
    git cherry-pick ${BRANCH} --strategy-option theirs
    echo "Pushing CHANGELOG.md to master..."
    git push origin HEAD:master
    git checkout $BRANCH
fi

echo "Pushing CHANGELOG.md to ${BRANCH}..."
git push origin HEAD:${BRANCH}