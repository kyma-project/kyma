FROM eu.gcr.io/kyma-project/kyma-operator:master-98e02519

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
