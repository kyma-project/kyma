FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang:go1.11

RUN go get golang.org/x/tools/cmd/goimports