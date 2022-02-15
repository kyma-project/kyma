FROM eu.gcr.io/kyma-project/incubator/develop/k8s-tools:20201023-5de446cf

LABEL source=git@github.com:kyma-project/kyma.git

COPY . /patch

ENTRYPOINT /patch/patch.sh
