FROM golang:1.12.10-alpine3.10 as builder

RUN apk update && apk add curl

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/components/service-binding-usage-controller
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /binding-usage-controller ./cmd/controller

FROM scratch

COPY --from=builder /usr/bin/curl /bin/curl
COPY --from=builder ./binding-usage-controller /root/binding-usage-controller

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/binding-usage-controller"]
