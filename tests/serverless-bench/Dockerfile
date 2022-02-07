FROM python:alpine3.15
WORKDIR /home/locust
RUN apk update && apk add bash curl jq miller git build-base linux-headers c-ares coreutils \
&& git clone https://github.com/locustio/locust.git --depth 1 --branch 2.6.0 locust \
&& cd locust \
&& pip install . \
&& cd ../ \
&& rm -rf locost \
&& apk del build-base linux-headers git

ADD locust.py run-benchmarks.sh /home/locust/
