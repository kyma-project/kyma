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
test=$(git tag -l '[0-9]*.[0-9]*.[0-9]*'| sort -r -V | grep '^[^-rc]*$'| head -n1)
echo "TEST:" $test
kyma_get_last_release_version_return_version=$(curl --silent --fail --show-error -H "Authorization: token ${BOT_GITHUB_TOKEN}" "https://api.github.com/repos/kyma-project/kyma/releases" | jq -r 'del( .[] | select( (.prerelease == true) or (.draft == true) )) | sort_by(.tag_name | split(".") | map(tonumber)) | .[-1].tag_name')
echo "old:" $kyma_get_last_release_version_return_version
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