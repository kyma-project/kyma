FROM alpine:3.8

RUN apk --no-cache add ca-certificates
RUN apk add --no-cache curl

COPY ./targz /usr/local/bin/targz
COPY ./indexbuilder /usr/local/bin/indexbuilder
COPY ./preupgrade /usr/local/bin/preupgrade

LABEL source=git@github.com:kyma-project/kyma.git

ENTRYPOINT ["tail", "-f", "/dev/null"]

