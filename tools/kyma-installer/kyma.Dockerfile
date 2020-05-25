FROM eu.gcr.io/kyma-project/kyma-operator:master-aeef8ce7

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
