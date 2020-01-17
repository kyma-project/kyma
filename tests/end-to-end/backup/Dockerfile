FROM golang:1.11-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup
WORKDIR ${SRC_DIR}

COPY . ${SRC_DIR}/

RUN CGO_ENABLED=0 GOOS=linux go test -c ./ -o /backup.test

FROM alpine:3.8
RUN apk --no-cache upgrade && apk --no-cache add curl

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /backup.test /backup.test
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup/licenses /licenses

ENTRYPOINT [ "/backup.test", "-test.v" ]
