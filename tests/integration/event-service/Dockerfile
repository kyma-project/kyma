FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20190930-d28d219 as builder

ENV SRC_DIR=/workspace/go/src/github.com/kyma-project/kyma/tests/integration/event-service

RUN mkdir -p $SRC_DIR
COPY . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/apitests

FROM alpine:3.10
RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder /workspace/go/src/github.com/kyma-project/kyma/tests/integration/event-service/apitests.test .
RUN chmod +x apitests.test
COPY --from=builder /workspace/go/src/github.com/kyma-project/kyma/tests/integration/event-service/entrypoint.sh .

ENTRYPOINT ./entrypoint.sh
