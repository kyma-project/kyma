FROM eu.gcr.io/kyma-project/kyma-operator:master-2ab108d3

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
