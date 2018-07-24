FROM alpine:3.7

LABEL source="git@github.com:kyma-project/kyma.git"

RUN apk update
RUN apk add python2
RUN apk add py2-flask

COPY . /istio-webhook
WORKDIR /istio-webhook
ENTRYPOINT ["python", "app.py"]
EXPOSE 5000
