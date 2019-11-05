FROM scratch

LABEL source = git@github.com:kyma-project/kyma.git

COPY ./licenses /app/licenses
COPY ./bin/asyncapi /app/asyncapi

ENTRYPOINT ["/app/asyncapi"]