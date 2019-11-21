FROM gcr.io/distroless/static:latest

LABEL source = git@github.com:kyma-project/kyma.git

COPY ./licenses /app/licenses
COPY ./bin/asyncapi /app/asyncapi

CMD ["/app/asyncapi"]