FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-registry-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -c ./test/apitests
RUN CGO_ENABLED=0 go test -c ./test/k8stests

FROM eu.gcr.io/kyma-project/external/alpine:3.12.0

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/apitests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/k8stests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-registry-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
