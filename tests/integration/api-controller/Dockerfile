FROM golang:1.9-alpine3.7 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/api-controller-acceptance-tests/
WORKDIR ${SRC_DIR}

COPY ./apicontroller ${SRC_DIR}/apicontroller/
COPY ./vendor ${SRC_DIR}/vendor/

RUN go test -c ./apicontroller/ -o /apicontroller.test

FROM alpine:3.7

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /apicontroller.test /apicontroller.test

ENTRYPOINT [ "/apicontroller.test", "-test.v" ]
