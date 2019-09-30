FROM eu.gcr.io/kyma-project/develop/kyma-operator:63f27f76

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
