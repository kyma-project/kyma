ARG INSTALLER_VERSION=55bc6038
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/develop
FROM $INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
