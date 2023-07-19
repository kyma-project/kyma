set -e
apk add nodejs npm
mkdir -p "/tmp/bin"
export PATH="/tmp/bin:${PATH}"
pushd "/tmp/bin" || exit
curl -sSLo kyma "https://storage.googleapis.com/kyma-cli-unstable/kyma-linux?alt=media"
chmod +x kyma
kyma_version=$(kyma version --client)
popd || exit
kyma provision k3d --ci
kyma_get_last_release_version_return_version=$(git tag -l '[0-9]*.[0-9]*.[0-9]*'| sort -r -V | grep '^[^-rc]*$'| head -n1)
export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
kyma deploy --ci --source ${KYMA_SOURCE} --timeout 60m
git reset --hard
git checkout tags/"${KYMA_SOURCE}"
make -C "./tests/fast-integration" "ci-pre-upgrade"
pushd "/tmp/bin" || exit
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/latest/download/kyma_linux_x86_64.tar.gz" && tar -zxvf kyma.tar.gz && chmod +x kyma && rm -f kyma.tar.gz
kyma_version=$(kyma version --client)
popd || exit
kyma deploy --ci --source "main" --timeout 20m
make -C "./tests/fast-integration" "ci-post-upgrade"
sleep 60
make -C "./tests/fast-integration" "ci-pre-upgrade"
