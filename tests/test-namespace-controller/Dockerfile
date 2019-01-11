FROM golang:1.10.2-alpine3.7 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/kyma/tests/test-namespace-controller/
WORKDIR ${BASE_APP_DIR}

# Copy files

COPY ./cmd/quantity/main.go ${BASE_APP_DIR}/cmd/quantity/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/

# Build app

RUN go build -o /quantity-to-int ./cmd/quantity/main.go

FROM alpine:3.7

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl
RUN apk add --no-cache unzip
RUN apk add --no-cache bash

## Install kubectl

RUN curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.8.4/bin/linux/amd64/kubectl
RUN chmod +x /usr/local/bin/kubectl

COPY ./test-namespace-controller.sh /test-namespace-controller.sh
COPY ./sample-namespace.yaml /sample-namespace.yaml
COPY --from=builder /quantity-to-int /quantity-to-int

CMD ["/test-namespace-controller.sh"]
