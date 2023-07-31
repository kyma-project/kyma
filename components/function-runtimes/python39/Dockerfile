FROM python:3.9.17-alpine3.17


# Serverless
LABEL source = git@github.com:kyma-project/kyma.git

# build-base and linux-headers are needed to install all requirements
RUN apk add --no-cache --virtual .build-deps build-base linux-headers

COPY kubeless/requirements.txt /kubeless/requirements.txt

RUN pip install --no-cache-dir -r /kubeless/requirements.txt

COPY kubeless/ /

WORKDIR /

USER 1000
# Tracing propagators are configured based on OTEL_PROPAGATORS env variable https://opentelemetry.io/docs/instrumentation/python/manual/#using-environment-variables
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3multi

CMD ["python", "/kubeless.py"]
