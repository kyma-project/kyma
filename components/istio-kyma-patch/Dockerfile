FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20201016-058ecbc7

LABEL source=git@github.com:kyma-project/kyma.git

COPY . /patch

ENTRYPOINT /patch/patch.sh
