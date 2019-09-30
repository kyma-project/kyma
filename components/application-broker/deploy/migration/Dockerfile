FROM alpine:3.8

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY ./ab-upgrade-protector /root/ab-upgrade-protector

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["/root/ab-upgrade-protector"]
