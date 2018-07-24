FROM alpine:3.7

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY ./helm-broker /root/helm-broker

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/helm-broker"]