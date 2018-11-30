FROM golang:1.9-alpine3.7 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/knative-serving-acceptance/
WORKDIR ${SRC_DIR}

COPY ./apicontroller ${SRC_DIR}/apicontroller/
COPY ./vendor ${SRC_DIR}/vendor/

RUN go test -c ./ -o /knative_serving.test

FROM alpine:3.7

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /knative_serving.test /knative_serving.test

ENTRYPOINT [ "/knative_serving.test", "-test.v" ]
