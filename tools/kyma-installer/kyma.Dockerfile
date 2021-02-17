ARG INSTALLER_VERSION="fb88f1a6"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project
FROM $INSTALLER_DIR/kyma-operator:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
