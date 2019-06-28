ARG INSTALLER_VERSION="408cb6a6"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/develop
FROM $INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
