ARG INSTALLER_VERSION=PR-2003
FROM eu.gcr.io/kyma-project/pr/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
