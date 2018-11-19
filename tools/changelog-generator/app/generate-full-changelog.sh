#!/usr/bin/env sh

source ${APP_PATH}/variables.sh

CONFIGURE_GIT=$1
if [ "$CONFIGURE_GIT" = "--configure-git" ]; then
    source ${APP_PATH}/config-git.sh
fi

source ${APP_PATH}/config-setup.sh

if [ ! -f $FULL_CHANGELOG_FILE_PATH ]; then
    FIRST_COMMIT=$(git rev-list --max-parents=0 HEAD)
    echo "Generating changelog starting from first commit '${FIRST_COMMIT}'..."
    lerna-changelog --from=${FIRST_COMMIT} | sed -e "s/## Unreleased/## ${NEW_RELEASE_TITLE}/g" > ${FULL_CHANGELOG_FILE_PATH}
else
    echo "Generating release changelog and prepending it to the CHANGELOG.md file..."
    if [ ! -f $RELEASE_CHANGELOG_FILE_PATH ]; then
        echo "ERROR: Generate release changelog first!"
        exit 1
    fi

    awk '/<!-- tocstop -->/ {p=1;next}p' ${FULL_CHANGELOG_FILE_PATH} > ${FULL_CHANGELOG_TEMP_FILE_PATH}
    printf "$(cat $RELEASE_CHANGELOG_FILE_PATH)\n\n$(cat ${FULL_CHANGELOG_TEMP_FILE_PATH})" > $FULL_CHANGELOG_FILE_PATH
    rm ${FULL_CHANGELOG_TEMP_FILE_PATH}
fi

echo "Generating navigation for CHANGELOG.md..."
printf "<!-- toc -->\n$(cat $FULL_CHANGELOG_FILE_PATH)" > $FULL_CHANGELOG_FILE_PATH
markdown-toc -i --maxdepth=2 $FULL_CHANGELOG_FILE_PATH

source ${APP_PATH}/config-cleanup.sh