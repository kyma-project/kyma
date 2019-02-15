ARG INSTALLER_VERSION=e48cc3ac
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/develop
FROM $INSTALLER_DIR/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
