FROM alpine:latest

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk add --no-cache curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.0/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

# To automatically get the latest version:
#RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

RUN curl -Lo /tmp/kubeless.zip https://github.com/kubeless/kubeless/releases/download/v1.0.0-alpha.7/kubeless_linux-amd64.zip && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

# To automatically get the latest version:
#RUN curl -Lo /tmp/kubeless.zip "$(curl -s https://api.github.com/repos/kubeless/kubeless/releases/latest | jq -r '.assets[]|select(.name=="kubeless_linux-amd64.zip").browser_download_url')" && unzip -q /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/bin/kubeless

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY ns.yaml k8s.yaml dependencies.json hello.js event.js svcbind.yaml /
COPY bin/app /test-kubeless

CMD ["/test-kubeless"]
