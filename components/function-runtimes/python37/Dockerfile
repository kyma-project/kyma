# kubeless/python:3.7
FROM kubeless/python@sha256:33aecef513c462d8cab3f18a4045be5fa4f8ca4617ec015ce40abe04a35ac6a0 as source

FROM python:3.7.8-alpine3.12

LABEL source = git@github.com:kyma-project/kyma.git

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger==0.4.6 prometheus_client==0.8.0

WORKDIR /
COPY --from=source /kubeless.py /kubeless.py

USER 1000

CMD ["python", "/kubeless.py"]