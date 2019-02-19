FROM golang:1.10-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test
WORKDIR ${SRC_DIR}

COPY restore ${SRC_DIR}/restore/

RUN go test -c ./restore/ -o /restore.test

FROM alpine:3.8
RUN apk --no-cache upgrade && apk --no-cache add curl

LABEL source = git@github.com:kyma-project/kyma.git
RUN mkdir -p /root/.kube && touch /root/.kube/config

COPY --from=builder /restore.test /restore.test
COPY restore/function.yaml /function.yaml

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
