ARG INSTALLER_VERSION="e029fcf4"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project/develop
FROM $INSTALLER_DIR/kyma-operator:63f27f76

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
