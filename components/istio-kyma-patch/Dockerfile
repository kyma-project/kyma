FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

ENV KUBECTL_VERSION 1.16.3

RUN apk --no-cache upgrade \
    && apk --no-cache --update add curl \
    && curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

COPY . /patch

RUN apk add bash

ENTRYPOINT /patch/patch.sh