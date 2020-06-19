FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200617-32c1f3ff

LABEL source=git@github.com:kyma-project/kyma.git

COPY . /patch

ENTRYPOINT /patch/patch.sh
