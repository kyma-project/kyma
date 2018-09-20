FROM alpine:3.7

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl
RUN apk add --no-cache bash

# Install kubectl
ARG KUBECTL_CLI_VERSION=1.9.0
RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v$KUBECTL_CLI_VERSION/bin/linux/amd64/kubectl
RUN chmod +x /usr/local/bin/kubectl
RUN /bin/bash -c 'printf "\n########################################\nkubectl version: $(kubectl version)\n########################################\n"'

ADD installer ./installer

COPY ../../resources /kyma/injected/resources
COPY ../../installation /kyma/injected/installation

ENTRYPOINT ["/installer"]
