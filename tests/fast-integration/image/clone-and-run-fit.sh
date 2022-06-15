#!/usr/bin/env ash

set -ex

# Clone the Kyma fast-integration tests and run the given make target

git clone https://github.com/kyma-project/kyma /kyma
cd /kyma/tests/fast-integration
make "$FIT_MAKE_TARGET"
