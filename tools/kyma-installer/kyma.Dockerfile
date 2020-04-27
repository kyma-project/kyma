FROM eu.gcr.io/kyma-project/kyma-operator:master-19fd6fa5
LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
