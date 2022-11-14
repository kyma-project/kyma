#!/usr/bin/env bash

readonly CURRENT_DIR=$1

echo "Hi1"
echo "$CURRENT_DIR"

ls ../telemetry-operator/config/crd/bases
echo "Hi"
ls $CURRENT_DIR

diff -q $CURRENT_DIR ../telemetry-operator/config/crd/bases