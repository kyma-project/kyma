FROM google/cloud-sdk:224.0.0-alpine

LABEL source="git@github.com:kyma-project/kyma.git"

ARG KUBECTL_CLI_VERSION
ARG SC_CLI_VERSION

RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v$KUBECTL_CLI_VERSION/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

RUN curl -L https://github.com/kyma-incubator/k8s-service-catalog/releases/download/v$SC_CLI_VERSION/service-catalog-installer-v$SC_CLI_VERSION-linux.tgz  | tar xz -C /usr/local/bin/ && \
    chmod +x /usr/local/bin/sc

COPY ./gcp-provider.sh /usr/local/bin/gcp-broker
RUN chmod +x /usr/local/bin/gcp-broker

