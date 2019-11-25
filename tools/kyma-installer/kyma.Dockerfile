ARG INSTALLER_VERSION="b2a53769"
ARG INSTALLER_DIR=eu.gcr.io/kyma-project
FROM eu.gcr.io/kyma-project/kyma-operator:b2a53769

LABEL source="git@github.com:kyma-project/kyma.git"

COPY /resources /kyma/injected/resources
