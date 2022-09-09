#!/usr/bin/env ash

set -ex

# Set the KYMA_VERSION to the git branch corresponding to the given release tag.
# For example: The "release-2.6" is the git branch corresponding to the "2.6.0" tag.
#if [[ -n $KYMA_VERSION && $KYMA_VERSION != "main" && $KYMA_VERSION != "latest" ]]; then
#  KYMA_VERSION=$(echo "$KYMA_VERSION" | awk -F '.' '{str = sprintf("release-%s.%s", $1, $2)} END {print str}')
#fi

# Clone the Kyma fast-integration tests and run the given make target with the given KYMA_VERSION.
#git clone --depth 1 --branch "$KYMA_VERSION" https://github.com/kyma-project/kyma /kyma
git clone --depth 1 --branch release-2.6 https://github.com/marcobebway/kyma /kyma

cd /kyma/tests/fast-integration
make "$FIT_MAKE_TARGET"
