FROM alpine:latest
LABEL source=git@github.com:kyma-project/kyma.git
WORKDIR /root

RUN apk update &&\
	apk add curl bash grep &&\
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&\
	chmod +x kubectl &&\
	mv kubectl /usr/local/bin/kubectl

ADD ./sar-test.sh sar-test.sh
ADD ./kyma-test-bindings.yaml kyma-test-bindings.yaml

ENTRYPOINT ["/bin/bash", "-c", "/root/sar-test.sh"]
