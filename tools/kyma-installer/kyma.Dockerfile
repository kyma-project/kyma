ARG INSTALLER_VERSION="PR-4684"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/pr
FROM $INSTALLER_DIR/kyma-operator:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
