FROM alpine:3.8
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /nats-streaming/

COPY prepare-config.sh .

ENTRYPOINT ["/nats-streaming/prepare-config.sh"]
