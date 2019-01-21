ARG INSTALLER_VERSION=PR-2376
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/pr
FROM $INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
