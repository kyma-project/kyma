FROM debian:stretch-slim
MAINTAINER jakub.blaszczyk@sap.com

WORKDIR /root

RUN apt-get update &&\
	apt-get install -y curl &&\
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&\
	chmod +x kubectl &&\
	mv kubectl /usr/local/bin/kubectl
LABEL source=git@github.com:kyma-project/kyma.git
ADD ./sar-test.sh sar-test.sh
ENTRYPOINT ["/bin/bash", "-c", "/root/sar-test.sh"]
