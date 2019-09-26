FROM golang:1.10-alpine as builder

ENV SRC_DIR /go/src/github.com/kyma-project/kyma/tests/tekton-pipelines/
WORKDIR ${SRC_DIR}

COPY ./tekton-tester.go ${SRC_DIR}/
COPY ./before-commit.sh ${SRC_DIR}/
COPY ./Gopkg.lock ${SRC_DIR}/
COPY ./Gopkg.toml ${SRC_DIR}/

RUN apk add bash dep git && ${SRC_DIR}/before-commit.sh
RUN go build -o /tekton_pipelines.test

FROM alpine:3.8

LABEL source = git@github.com:kyma-project/kyma.git

COPY --from=builder /tekton_pipelines.test /tekton_pipelines.test

ENTRYPOINT [ "/tekton_pipelines.test", "-test.v" ]
