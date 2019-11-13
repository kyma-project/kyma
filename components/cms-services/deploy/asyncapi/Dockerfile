# Get latest CA certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY ./licenses /app/licenses
COPY ./bin/asyncapi /app/asyncapi
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/asyncapi"]