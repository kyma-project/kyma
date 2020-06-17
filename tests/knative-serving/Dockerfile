FROM golang:1.14-alpine3.12 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/knative-serving/
WORKDIR ${SRC_DIR}

COPY ./knative_serving_test.go ${SRC_DIR}/
COPY ./before-commit.sh ${SRC_DIR}/
COPY ./Gopkg.lock ${SRC_DIR}/
COPY ./Gopkg.toml ${SRC_DIR}/
COPY ./licenses/ ./licenses

RUN apk add bash dep git build-base && ${SRC_DIR}/before-commit.sh
RUN go test -c ./ -o /knative_serving.test

FROM alpine:3.12

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /knative_serving.test /knative_serving.test
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/knative-serving/licenses /licenses

ENTRYPOINT [ "/knative_serving.test", "-test.v" ]
