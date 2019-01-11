FROM alpine:3.8

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk add --no-cache curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.10.0/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

RUN mkdir -p /root/.kube && touch /root/.kube/config

COPY testCounterPod.yaml /
COPY bin/app /test-logging

CMD ["/test-logging"]
