FROM golang:1.9-alpine3.7 as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test
WORKDIR ${SRC_DIR}

COPY restore ${SRC_DIR}/restore/

RUN go test -c ./restore/ -o /restore.test
RUN apk --no-cache upgrade && apk --no-cache add curl
RUN curl -Lo /tmp/ark.tgz https://github.com/heptio/ark/releases/download/v0.10.1/ark-v0.10.1-linux-amd64.tar.gz && tar xf /tmp/ark.tgz -C /tmp
RUN curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.0/bin/linux/amd64/kubectl && chmod +x /tmp/kubectl 
FROM alpine:3.7

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /restore.test /restore.test
COPY --from=builder /tmp/ark /usr/bin/ark
COPY --from=builder /tmp/kubectl /usr/bin/kubectl
COPY restore/function.yaml /function.yaml
COPY restore/gateway.yaml /gateway.yaml

RUN mkdir -p /root/.kube && touch /root/.kube/config
ENTRYPOINT [ "/restore.test", "-test.v" ]
