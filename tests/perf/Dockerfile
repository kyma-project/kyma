FROM alpine

COPY ./ /etc/tests
RUN chmod +x /etc/tests/prerequisites/ory/setup.sh

RUN  apk --no-cache --update add curl \
    && curl -L https://github.com/loadimpact/k6/releases/download/v0.26.0/k6-v0.26.0-linux32.tar.gz -o k6.tar.gz \
    && tar -xvzf k6.tar.gz \
    && mv k6-v0.26.0-linux32/k6 /usr/local/bin/k6 \
    && chmod +x /usr/local/bin/k6 \
    && rm k6.tar.gz

RUN curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.14.7/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

ENTRYPOINT ["/bin/bash"]