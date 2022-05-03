FROM debian:11

ENV KUBECONFIG=/KUBECONFIG/kubeconfig.yaml
ENV EVENTMESH_SECRET_FILE=/EMS/ems.json

VOLUME /KUBECONFIG
VOLUME /EMS

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    curl \
    git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fLSs https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get -y --no-install-recommends install nodejs

WORKDIR /
COPY clone-and-run-eventing-fit.sh /run-fit.sh
RUN chmod +x /run-fit.sh

CMD ["./run-fit.sh"]
