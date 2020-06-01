FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200507-070ff576

LABEL source=git@github.com:kyma-project/kyma.git

COPY . /patch

ENTRYPOINT /patch/patch.sh
