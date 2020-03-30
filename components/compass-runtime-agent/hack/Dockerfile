FROM golang:1.12

ENV GO111MODULE=off

WORKDIR /go/src/k8s.io

RUN git clone https://github.com/kubernetes/code-generator.git
# Versions needs to be compatible with Kuberentes client used (currently Kubebuilder requires 1.14.*)
RUN cd code-generator && git checkout release-1.17

RUN go get k8s.io/apimachinery; exit 0

ARG repo="${GOPATH}/src/${PROJECT_MODULE}"

RUN mkdir -p $repo

WORKDIR $GOPATH/src/k8s.io/code-generator

ENV GO111MODULE=on

RUN go mod vendor

ENV GO111MODULE=off

VOLUME $repo