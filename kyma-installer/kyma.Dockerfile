ARG INSTALLER_VERSION=7bd3337d
ARG INSTALLER_DIR=develop
FROM eu.gcr.io/kyma-project/$INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
