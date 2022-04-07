#!/usr/bin/env bash

kubectl -n kyma-system patch service monitoring-alertmanager --type=json -p='[{"op": "remove", "path": "/spec/selector/app"}]'