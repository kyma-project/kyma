FROM alpine:3.13
LABEL source=git@github.com:kyma-project/kyma.git

RUN apk update && \
    ## Solves CVE-2020-1971
    apk add --no-cache openssl=1.1.1j-r0

WORKDIR /nats-streaming/

COPY scripts/prepare-config.sh .

ENTRYPOINT ["/nats-streaming/prepare-config.sh"]
