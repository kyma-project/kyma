FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/service-catalog
ADD . $SRC_DIR
COPY ./licenses/ $SRC_DIR/licenses/
COPY entrypoint.sh /

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /env-tester.bin ./cmd/env-tester/main.go
RUN CGO_ENABLED=0 go test -tags=acceptance -c ./test -o /servicecatalog.test

# this image is not based on scratch because it needs
# CURL and shell to remove istio sidecar via quitquitquit endpoint
FROM alpine:3.10
RUN apk add --no-cache curl

COPY --from=builder env-tester.bin /go/bin/env-tester.bin
COPY --from=builder ./servicecatalog.test .
COPY ./licenses/ /app/licenses
COPY --from=builder /entrypoint.sh .

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ./entrypoint.sh
