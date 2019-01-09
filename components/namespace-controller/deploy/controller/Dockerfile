FROM alpine:3.7

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

LABEL source="github.com:kyma-project/kyma.git"

ADD namespace-controller /namespace-controller

ENTRYPOINT ["/namespace-controller"]
