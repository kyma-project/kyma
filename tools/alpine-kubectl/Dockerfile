FROM alpine:3.8

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk update &&\
	apk add openssl coreutils curl bash &&\
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl &&\
	chmod +x ./kubectl &&\
	mv ./kubectl /usr/local/bin/kubectl

ENTRYPOINT ["/bin/bash"]
