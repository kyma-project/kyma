FROM eu.gcr.io/kyma-project/kyma-operator:77306d33

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
