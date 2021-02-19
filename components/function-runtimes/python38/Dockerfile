# kubeless/python:3.7
# Temporary image built manually to contain security patches for vulnerability CVE-2021-3177,
# replace with official python image once python 3.8.8 is released
FROM eu.gcr.io/kyma-project/incubator/python:20210219-652db865

LABEL source = git@github.com:kyma-project/kyma.git

RUN pip install bottle==0.12.13 cherrypy==8.9.1 wsgi-request-logger==0.4.6 prometheus_client==0.8.0

COPY kubeless/ /

WORKDIR /

USER 1000

CMD ["python", "/kubeless.py"]
