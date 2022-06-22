FROM alpine:3.14.6

ENV FIT_MAKE_TARGET=ci-skr

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

RUN apk --no-cache add \
    build-base \
    git \
    curl \
    nodejs=14.19.0-r0 \
    npm

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
