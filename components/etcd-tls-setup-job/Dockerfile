FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200309-7eaac6d8

LABEL source=git@github.com:kyma-project/kyma.git

ENV CFSSL_VERSION 1.2

RUN apk --no-cache upgrade \
    && curl -o /usr/local/bin/cfssl https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssl_linux-amd64 && chmod +x /usr/local/bin/cfssl \
    && curl -o /usr/local/bin/cfssljson https://pkg.cfssl.org/R${CFSSL_VERSION}/cfssljson_linux-amd64 && chmod +x /usr/local/bin/cfssljson

ADD ./bin /etcd-tls-setup/bin
ADD ./config /etcd-tls-setup/config

ENTRYPOINT ["/bin/sh"]