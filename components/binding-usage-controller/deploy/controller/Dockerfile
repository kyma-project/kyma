FROM alpine:3.7

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY ./binding-usage-controller /root/binding-usage-controller

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/binding-usage-controller"]