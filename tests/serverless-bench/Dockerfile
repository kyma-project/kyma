FROM python:alpine3.15
WORKDIR /home/locust
RUN apk add --no-cache bash curl jq miller git build-base linux-headers c-ares coreutils

COPY requirements.txt /home/locust/
RUN pip3 install --no-cache-dir -r requirements.txt

COPY locust.py run-benchmarks.sh /home/locust/
