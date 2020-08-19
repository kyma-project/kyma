FROM golang:1.14-alpine3.12 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/kyma/tests/integration/monitoring
ADD . $SRC_DIR

WORKDIR $SRC_DIR

RUN CGO_ENABLED=0 go build -o /test-monitoring

FROM alpine:3.12

RUN apk --no-cache upgrade && apk --no-cache add ca-certificates curl

RUN curl -Lo /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.16.12/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl

# Variables used for labeling created images
LABEL source=git@github.com:kyma-project/kyma.git

RUN mkdir -p /root/.kube && touch /root/.kube/config
COPY --from=builder /test-monitoring /test-monitoring
COPY ./licenses/ /licenses

ENTRYPOINT ["/test-monitoring"]
