FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20190325-ff66a3a

ARG KYMA_VERSION

ENV KYMA_VERSION ${KYMA_VERSION}
ENV TILLER_RESOURCE "tiller.yaml"
ENV KYMA_CONFIG_RESOURCE "kyma-config-cluster.yaml"
ENV KYMA_INSTALLER_RESOURCE "kyma-installer-cluster.yaml"

RUN mkdir -p /deployer/scripts
WORKDIR /deployer

RUN curl https://raw.githubusercontent.com/kyma-project/kyma/${KYMA_VERSION}/installation/resources/tiller.yaml --output ${TILLER_RESOURCE} --fail -L \
    && curl https://github.com/kyma-project/kyma/releases/download/${KYMA_VERSION}/kyma-config-cluster.yaml --output ${KYMA_CONFIG_RESOURCE} --fail -L \
    && curl https://github.com/kyma-project/kyma/releases/download/${KYMA_VERSION}/kyma-installer-cluster.yaml --output ${KYMA_INSTALLER_RESOURCE} --fail -L


COPY scripts/*  scripts/

ENTRYPOINT ["scripts/deploy.sh"]
