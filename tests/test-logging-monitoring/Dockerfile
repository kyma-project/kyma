FROM debian:stable-slim

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apt-get update && apt-get install -y curl unzip && rm -r /var/lib/apt/lists/*

RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.9.0/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY bin/app /test-logging-monitoring

CMD ["/test-logging-monitoring"]
