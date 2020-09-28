FROM eu.gcr.io/kyma-project/kyma-operator:4260fe58
LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
