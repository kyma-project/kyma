FROM eu.gcr.io/kyma-project/kyma-operator:45473210

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/
