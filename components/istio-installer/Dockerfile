FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200121-4f3202bd

LABEL source=git@github.com:kyma-project/kyma.git

ENV ISTIOCTL_VERSION 1.4.4

RUN apk add bash curl &&\
	curl -L https://github.com/istio/istio/releases/download/${ISTIOCTL_VERSION}/istioctl-${ISTIOCTL_VERSION}-linux.tar.gz -o istioctl.tar.gz &&\
	tar xvzf istioctl.tar.gz &&\
	chmod +x istioctl &&\
	mv istioctl /usr/local/bin/istioctl &&\
	rm istioctl.tar.gz

ENTRYPOINT ["/bin/bash"]
