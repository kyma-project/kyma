FROM alpine:3.7

LABEL source="git@github.com:kyma-project/kyma.git"
ADD /bin/app /app
COPY ./licenses/ licenses/

ENTRYPOINT [ "/app"]