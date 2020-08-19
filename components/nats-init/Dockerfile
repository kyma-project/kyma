FROM alpine:3.12
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /nats-streaming/

COPY scripts/prepare-config.sh .

ENTRYPOINT ["/nats-streaming/prepare-config.sh"]
