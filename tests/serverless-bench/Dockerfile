FROM python:alpine3.15
WORKDIR /home/locust
RUN apk update && apk add bash curl jq miller git build-base linux-headers c-ares coreutils

ADD requirements.txt /home/locust/
RUN pip3 install -r requirements.txt

ADD locust.py run-benchmarks.sh /home/locust/
