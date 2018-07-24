FROM debian:stable-slim

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apt-get update && apt-get install -y curl unzip && rm -r /var/lib/apt/lists/*

RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.9.0/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

RUN curl -Lo /tmp/kubeless.zip https://github.com/kubeless/kubeless/releases/download/v1.0.0-alpha.2/kubeless_linux-amd64.zip && unzip -uq /tmp/kubeless.zip -d /tmp/ && mv /tmp/bundles/kubeless_linux-amd64/kubeless /usr/local/bin/ && rm -r /tmp/kubeless.zip /tmp/bundles && chmod +x /usr/local/bin/kubeless

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY bin/app /test-kubeless
COPY ingress.yaml route.yaml hello.js dependecies.json /

CMD ["/test-kubeless"]
