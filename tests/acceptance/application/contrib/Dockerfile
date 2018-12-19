FROM alpine:3.8

RUN apk update && apk add curl

ADD client.bin /go/bin/client.bin
ADD gateway.bin /go/bin/gateway.bin
ADD re.test /go/bin/re.test

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT /go/bin/tester.bin