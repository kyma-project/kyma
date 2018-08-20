#!/usr/bin/env bash
BASEDIR=$(dirname "$0")

cp $BASEDIR/../../../docs/application-connector/docs/assets/connectorapi.yaml ../

(cd $BASEDIR/../ ; docker build . "${@:1}")

rm $BASEDIR/../connectorapi.yaml