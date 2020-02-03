FROM eu.gcr.io/kyma-project/kyma-operator:master-e0c99984

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
