FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

ENV KUBECTL_VERSION 1.10.6

RUN apk --no-cache upgrade \
    && apk --no-cache --update add curl \
    && curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl \
    && curl -o /serving.yaml https://github.com/knative/serving/releases/download/v0.2.1/release-no-mon.yaml \
    && curl -o /eventing.yaml https://github.com/knative/eventing/releases/download/v0.2.0/release.yaml

COPY . /app

RUN apk add bash

ENTRYPOINT /app/install.sh