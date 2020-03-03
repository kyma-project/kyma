FROM golang:1.13-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-operator-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 go test -o applicationcontroller.test -c ./test/applicationcontroller/tests
RUN CGO_ENABLED=0 go test -o serviceinstancecontroller.test -c ./test/serviceinstancecontroller/tests
RUN CGO_ENABLED=0 go test -o applicationtests.test -c ./test/application/tests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/applicationcontroller.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/serviceinstancecontroller.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/applicationtests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
