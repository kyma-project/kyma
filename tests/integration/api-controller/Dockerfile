FROM golang:1.12-alpine3.10 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/integration/api-controller/

WORKDIR ${SRC_DIR}

COPY ./apicontroller ${SRC_DIR}/apicontroller/
COPY ./vendor ${SRC_DIR}/vendor/

RUN CGO_ENABLED=0 go test -c ./apicontroller/ -o /apicontroller.test

FROM alpine:3.10

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /apicontroller.test /apicontroller.test

ENTRYPOINT [ "/apicontroller.test", "-test.v" ]
