# syntax=docker/dockerfile:1
FROM ubuntu:18.04
COPY . /app
RUN ["chmod", "+x", "/app/constant-logging-script.sh"]
ENTRYPOINT [ "sh", "/app/constant-logging-script.sh" ]
