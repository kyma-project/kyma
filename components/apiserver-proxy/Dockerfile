FROM alpine:3.8

LABEL source="git@github.com:kyma-project/kyma.git"
ADD /bin/app /app
COPY ./licenses/ licenses/

ENTRYPOINT [ "/app"]