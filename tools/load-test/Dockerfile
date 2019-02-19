FROM alpine:3.8
LABEL source="git@github.com:kyma-project/kyma.git"
RUN apk --no-cache upgrade && apk --no-cache add curl
RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.11.0/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl
# To automatically get the latest version:
#RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl
RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY k8syaml/ns.yaml k8syaml/ns.yaml
COPY k8syaml/function.yaml k8syaml/function.yaml
COPY bin/app /load-test
CMD ["/load-test"]
