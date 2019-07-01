FROM alpine:3.8

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

ADD kyma-operator ./kyma-operator

ENTRYPOINT ["/kyma-operator"]
