FROM alpine:3.10

LABEL source = git@github.com:kyma-project/kyma.git

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY ./licenses /app/licenses
COPY ./bin/asyncapi /app/asyncapi

CMD ["/app/asyncapi"]