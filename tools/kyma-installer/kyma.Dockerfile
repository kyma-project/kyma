ARG INSTALLER_VERSION="c951343e"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project
FROM $INSTALLER_DIR/kyma-operator:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
