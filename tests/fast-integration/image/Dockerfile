FROM debian:11

ENV KUBECONFIG=/KUBECONFIG/kubeconfig.yaml
ENV EVENTMESH_SECRET_FILE=/EMS/ems.json
ENV FIT_MAKE_TARGET=ci-skr

VOLUME /KUBECONFIG
VOLUME /EMS

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

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
COPY clone-and-run-fit.sh /clone-and-run-fit.sh
RUN chmod +x /clone-and-run-fit.sh

CMD ["./clone-and-run-fit.sh"]
