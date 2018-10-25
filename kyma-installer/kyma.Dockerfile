ARG INSTALLER_VERSION=2dc3ac03

FROM eu.gcr.io/kyma-project/develop/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected/installation
COPY /resources /kyma/injected/resources
