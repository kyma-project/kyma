#!/usr/bin/env bash

## This script copies charts from Compass repository
## Make sure Compass repository location is $GOPATH/src/github.com/kyma-incubator/compass
## And you are on a proper branch (that you want to copy charts from)

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

COMPASS_CHART_PATH=${GOPATH}/src/github.com/kyma-incubator/compass/chart/compass
COMPASS_CHARTS_PATH="${COMPASS_CHART_PATH}/charts"

PROVISIONER_CHART_PATH="${COMPASS_CHARTS_PATH}/provisioner"
echo "Copying Provisioner chart from ${PROVISIONER_CHART_PATH}"
cp -rf ${PROVISIONER_CHART_PATH}/* ${DIR}/
echo "Done"

echo "------------------------------------------------------------------"
echo "Remember to remove any deleted files and adjust root chart values!"
echo "------------------------------------------------------------------"
