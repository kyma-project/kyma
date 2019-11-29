FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tools/health-check-proxy
ADD . $SRC_DIR
WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /health-check ./cmd/health-check

FROM scratch

LABEL source=git@github.com:kyma-project/kyma.git

COPY --from=builder ./health-check /health-check