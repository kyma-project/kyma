FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

ENV KUBECTL_VERSION 1.10.6

RUN apk --no-cache upgrade \
    && apk --no-cache --update add curl \
    && curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl \
    && apk --no-cache add bash openssl

COPY . /app

ENTRYPOINT [ "/app/xip-patch.sh" ]