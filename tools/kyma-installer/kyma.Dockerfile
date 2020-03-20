FROM eu.gcr.io/kyma-project/kyma-operator:master-c74d3125
LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
