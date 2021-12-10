FROM golang:1.12.10-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/end-to-end/upgrade
COPY . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /e2e-upgrade-test ./main.go

FROM alpine:3.10

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY --from=builder ./e2e-upgrade-test /usr/local/bin/e2e-upgrade-test
COPY ./licenses /app/licenses

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["e2e-upgrade-test"]
