FROM google/cloud-sdk:225.0.0-alpine

LABEL source="git@github.com:kyma-project/kyma.git"

ARG KUBECTL_CLI_VERSION=v1.10.0
ARG SC_CLI_VERSION=v1.0.0-beta.5

RUN apk add coreutils

RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_CLI_VERSION/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl

RUN curl -L https://github.com/kyma-incubator/k8s-service-catalog/releases/download/$SC_CLI_VERSION/service-catalog-installer-$SC_CLI_VERSION-linux.tgz  | tar xz -C /usr/local/bin/ && \
    chmod +x /usr/local/bin/sc

COPY bin/gcp-broker.sh /usr/local/bin/gcp-broker
RUN chmod +x /usr/local/bin/gcp-broker

COPY bin/status-checker.sh /usr/local/bin/status-checker
RUN chmod +x /usr/local/bin/status-checker
