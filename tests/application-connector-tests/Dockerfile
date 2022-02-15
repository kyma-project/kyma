FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-connector-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -o applicationaccess.test -c ./test/applicationaccess/tests

FROM eu.gcr.io/kyma-project/external/alpine:3.12.0

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-connector-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-connector-tests/applicationaccess.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-connector-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
