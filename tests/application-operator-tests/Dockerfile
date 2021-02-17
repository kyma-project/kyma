FROM eu.gcr.io/kyma-project/external/golang:1.14.8-alpine as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/application-operator-tests
WORKDIR $SRC_DIR

COPY . $SRC_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN CGO_ENABLED=0 go test -o applicationcontroller.test -c ./test/applicationcontroller/tests
RUN CGO_ENABLED=0 go test -o serviceinstancecontroller.test -c ./test/serviceinstancecontroller/tests
RUN CGO_ENABLED=0 go test -o applicationtests.test -c ./test/application/tests

FROM eu.gcr.io/kyma-project/external/alpine:3.12.0

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/applicationcontroller.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/serviceinstancecontroller.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/applicationtests.test .
COPY --from=builder /go/src/github.com/kyma-project/kyma/tests/application-operator-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
