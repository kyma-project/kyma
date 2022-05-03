#!/usr/bin/env bash

set -e

git clone https://github.com/kyma-project/kyma /kyma
cd /kyma/tests/fast-integration
make ci-test-eventing