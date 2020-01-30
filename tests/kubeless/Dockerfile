FROM golang:1.10-alpine as builder
LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk --no-cache upgrade && apk --no-cache add curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.16.3/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

# To automatically get the latest version:
#RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

RUN curl -Lo /tmp/kubeless.zip https://github.com/kubeless/kubeless/releases/download/v1.0.5/kubeless_linux-amd64.zip && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/kubeless/

WORKDIR ${SRC_DIR}

COPY . ${SRC_DIR}/

RUN go generate ./... && \
    CGO_ENABLED=0 go build -o /kubeless-tests


# To automatically get the latest version:
#RUN curl -Lo /tmp/kubeless.zip "$(curl -s https://api.github.com/repos/kubeless/kubeless/releases/latest | jq -r '.assets[]|select(.name=="kubeless_linux-amd64.zip").browser_download_url')" && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

FROM alpine:3.10

RUN apk add --no-cache curl

COPY --from=builder /usr/bin/kubectl /usr/bin/kubectl
COPY --from=builder /usr/bin/kubeless /usr/bin/kubeless

RUN chmod +x /usr/bin/kubeless
RUN chmod +x /usr/bin/kubectl

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY ns.yaml hello.js package.json /
COPY --from=builder /kubeless-tests /kubeless-tests

CMD ["/kubeless-tests"]
