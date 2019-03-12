ARG INSTALLER_VERSION=0.8.0-rc1
ARG INSTALLER_DIR=eu.gcr.io/kyma-project
FROM $INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
