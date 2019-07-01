FROM alpine:3.9
LABEL source=git@github.com:kyma-project/kyma.git
WORKDIR /root

RUN apk update &&\
	apk add curl bash grep &&\
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&\
	chmod +x kubectl &&\
	mv kubectl /usr/local/bin/kubectl

ADD ./fetch-token/bin/app ./app
ADD ./test.sh ./test.sh
ENTRYPOINT ["/bin/bash", "-c", "/root/test.sh"]
