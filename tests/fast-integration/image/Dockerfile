FROM node:14.20.0-alpine3.16

ENV KYMA_VERSION=main
ENV FIT_MAKE_TARGET=ci-skr

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

RUN apk update; apk upgrade; rm /var/cache/apk/*

RUN apk --no-cache add \
    build-base \
    git \
    curl

#################################################################
################ Install kcp-cli ################################
#################################################################

# Note: kcp-cli doesn't provide releases
ENV KCPCLI_VERSION=master-1208b417
RUN curl -fLSs -o /usr/local/bin/kcp https://storage.googleapis.com/kyma-development-artifacts/kcp/${KCPCLI_VERSION}/kcp-linux &&\
    chmod +x /usr/local/bin/kcp

WORKDIR /
COPY clone-and-run-fit.sh /clone-and-run-fit.sh
RUN chmod +x /clone-and-run-fit.sh

CMD ["./clone-and-run-fit.sh"]
