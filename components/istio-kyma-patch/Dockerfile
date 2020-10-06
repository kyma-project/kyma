FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20201006-f331d22b

LABEL source=git@github.com:kyma-project/kyma.git

COPY . /patch

ENTRYPOINT /patch/patch.sh
