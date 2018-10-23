FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

ENV KUBECTL_VERSION 1.10.5
ENV CFSSL_VERSION 1.2

RUN apk --no-cache upgrade \
    && apk --no-cache --update add curl \
    && curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl \
    && curl -o /usr/local/bin/cfssl https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssl_linux-amd64 && chmod +x /usr/local/bin/cfssl \
    && curl -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssljson_linux-amd64 && chmod +x /usr/local/bin/cfssljson

ADD ./bin /etcd-tls-setup/bin
ADD ./config /etcd-tls-setup/config

ENTRYPOINT ["/bin/sh"]