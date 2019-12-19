FROM golang:1.11-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test
WORKDIR ${SRC_DIR}

COPY . ${SRC_DIR}/

RUN CGO_ENABLED=0 GOOS=linux go test -c ./ -o /restore.test

FROM alpine:3.8
RUN apk --no-cache upgrade && apk --no-cache add curl

LABEL source = git@github.com:kyma-project/kyma.git
RUN mkdir -p /root/.kube && touch /root/.kube/config

COPY --from=builder /restore.test /restore.test
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/licenses /licenses

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
