FROM eu.gcr.io/kyma-project/kyma-operator:5a771fdf

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
