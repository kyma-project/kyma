ARG INSTALLER_VERSION=b808b509
FROM eu.gcr.io/kyma-project/prow/test/develop/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
