FROM golang:1.9-alpine3.7 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test
WORKDIR ${SRC_DIR}

COPY restore ${SRC_DIR}/restore/
COPY ./vendor ${SRC_DIR}/vendor/

RUN go test -c ./restore/ -o /restore.test

FROM alpine:3.7

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /restore.test /restore.test

ENTRYPOINT [ "/restore.test", "-test.v" ]
