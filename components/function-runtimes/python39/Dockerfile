# image base on python:3.9.17-alpine3.18 (python:3.9.17, alpine:3.18.2)
FROM python@sha256:281811c93dbf7b29719709c129230edfa1b10585d720f95c60363a358e49b62c

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
ENV OTEL_PROPAGATORS=tracecontext,baggage,b3

CMD ["python", "/kubeless.py"]
