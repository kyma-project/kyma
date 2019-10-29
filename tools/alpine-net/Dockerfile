FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache upgrade && apk --no-cache --update add net-tools \
    bind-tools \
    curl \
    nmap

ENTRYPOINT ["/bin/sh"]