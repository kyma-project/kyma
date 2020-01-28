FROM golang:1.10-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/knative-eventing-channel-kafka/
WORKDIR ${SRC_DIR}

COPY ./kafka_channel_test.go ${SRC_DIR}/
COPY ./Gopkg.lock ${SRC_DIR}/
COPY ./Gopkg.toml ${SRC_DIR}/
COPY ./licenses/ ./licenses

RUN apk add bash dep git
RUN dep ensure -v --vendor-only && go test -c ./ -o /kafka_channel.test

FROM alpine:3.8

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /kafka_channel.test /kafka_channel.test
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/knative-eventing-channel-kafka/licenses /licenses

ENTRYPOINT [ "/kafka_channel.test", "-test.v" ]
