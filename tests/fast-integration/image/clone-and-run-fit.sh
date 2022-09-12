#!/usr/bin/env ash

set -ex

# Clone the Kyma fast-integration tests and run the given make target

git clone https://github.com/mfaizanse/kyma.git /kyma

cd /kyma
git checkout skr_provision_stage
git status

cd /kyma/tests/fast-integration
make "$FIT_MAKE_TARGET"
