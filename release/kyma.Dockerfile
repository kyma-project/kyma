ARG INSTALLER_VERSION

FROM eu.gcr.io/kyma-project/develop/installer:$INSTALLER_VERSION

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /installation /kyma/injected
COPY /resources /kyma/injected