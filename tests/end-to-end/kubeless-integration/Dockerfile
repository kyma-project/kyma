FROM golang:1.10 as builder

WORKDIR /go/src/github.com/kyma-project/kyma/tests/end-to-end/kubeless-integration
COPY vendor ./vendor
COPY test-kubeless.go .

RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o test-kubeless

FROM alpine:3.8

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk add --no-cache curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.15.2/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

# To automatically get the latest version:
#RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

RUN curl -Lo /tmp/kubeless.zip https://github.com/kubeless/kubeless/releases/download/v1.0.0/kubeless_linux-amd64.zip && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

# To automatically get the latest version:
#RUN curl -Lo /tmp/kubeless.zip "$(curl -s https://api.github.com/repos/kubeless/kubeless/releases/latest | jq -r '.assets[]|select(.name=="kubeless_linux-amd64.zip").browser_download_url')" && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY ns.yaml hello.js event.js svc-instance.yaml svc-binding.yaml /
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/end-to-end/kubeless-integration/test-kubeless /test-kubeless

CMD ["/test-kubeless"]
