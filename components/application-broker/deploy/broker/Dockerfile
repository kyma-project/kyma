FROM alpine:3.8

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

# Variables used for labeling created images
LABEL source=git@github.com:kyma-project/kyma.git

COPY ./application-broker /root/application-broker

ENTRYPOINT ["/root/application-broker"]