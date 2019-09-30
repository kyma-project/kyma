FROM eu.gcr.io/kyma-project/develop/kyma-operator:e029fcf4

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
