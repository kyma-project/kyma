#!/bin/bash
set -e

wget "${APP_REPOSITORY_URL}" --output-document sources.tar.gz
tar -vx --file sources.tar.gz --directory "${APP_MOUNT_PATH}"
