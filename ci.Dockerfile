FROM ubuntu:16.04

LABEL source="git@github.com:kyma-project/kyma.git"

ARG KUBECTL_CLI_VERSION
ARG KUBELESS_CLI_VERSION
ARG MINIKUBE_VERSION
ARG HELM_VERSION

# Get dependencies for curl of the docker
RUN apt-get update && apt-get install -y \
    bash \
    curl \
    jq \
    socat \
    sudo \
    vim \
    zip \
    && rm -rf /var/lib/apt/lists/*

# Install kubectl
RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v$KUBECTL_CLI_VERSION/bin/linux/amd64/kubectl
RUN chmod +x /usr/local/bin/kubectl

# Install Kubeless CLI
RUN curl -Lo /tmp/kubeless-binary.zip https://github.com/kubeless/kubeless/releases/download/v$KUBELESS_CLI_VERSION/kubeless_linux-amd64.zip && \
    unzip -uq /tmp/kubeless-binary.zip -d /tmp/ && \
    chmod +x /tmp/bundles/kubeless_linux-amd64/kubeless && \
    sudo mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/local/bin/ && \
    rm -rf /tmp/bundles && \
    rm -rf /tmp/kubeless-binary.zip

# Install Minikube
RUN curl -Lo /usr/local/bin/minikube https://storage.googleapis.com/minikube/releases/v$MINIKUBE_VERSION/minikube-linux-amd64
RUN chmod +x /usr/local/bin/minikube

# Install Docker from Docker Inc. repositories.
RUN curl -sSL https://get.docker.com/ | sh

# Install Helm
RUN curl -Lo /tmp/helm-linux-amd64.tar.gz https://kubernetes-helm.storage.googleapis.com/helm-v$HELM_VERSION-linux-amd64.tar.gz
RUN tar -xvf /tmp/helm-linux-amd64.tar.gz -C /tmp/
RUN chmod +x  /tmp/linux-amd64/helm && sudo mv /tmp/linux-amd64/helm /usr/local/bin/

# Copying into the container all the necessary files like scripts and resources definition
RUN mkdir /kyma

COPY . /kyma

ENV IGNORE_TEST_FAIL="true"
ENV RUN_TESTS="true"

RUN echo 'alias kc="kubectl"' >> ~/.bashrc

# minikube and docker start must be done on starting container to make it work
ENTRYPOINT /kyma/installation/scripts/docker-start.sh \
    && /kyma/installation/scripts/minikube.sh --vm-driver none \
    && /kyma/installation/scripts/installer-ci-local.sh \
    && /kyma/installation/scripts/is-installed.sh \
    && /kyma/installation/scripts/watch-pods.sh \
    && (($RUN_TESTS && /kyma/installation/scripts/testing.sh) || $IGNORE_TEST_FAIL) \
    && exec bash