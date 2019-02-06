FROM golang:1.10-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test
WORKDIR ${SRC_DIR}

COPY restore ${SRC_DIR}/restore/

RUN go test -c ./restore/ -o /restore.test
RUN apk --no-cache upgrade && apk --no-cache add curl
RUN curl -Lo /tmp/ark.tgz https://github.com/heptio/ark/releases/download/v0.10.1/ark-v0.10.1-linux-amd64.tar.gz && tar xf /tmp/ark.tgz -C /tmp
RUN curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.0/bin/linux/amd64/kubectl && chmod +x /tmp/kubectl 

FROM alpine:3.8
RUN apk --no-cache upgrade && apk --no-cache add curl iputils

LABEL source = git@github.com:kyma-project/kyma.git
RUN mkdir -p /root/.kube && touch /root/.kube/config

COPY --from=builder /tmp/ark /usr/bin/ark
COPY --from=builder /tmp/kubectl /usr/bin/kubectl
COPY --from=builder /restore.test /restore.test
COPY restore/function.yaml /function.yaml
COPY restore/gateway.yaml /gateway.yaml

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
