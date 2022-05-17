FROM alpine:latest

ENV KUBECONFIG=/KUBECONFIG/kubeconfig.yaml
ENV EVENTMESH_SECRET_FILE=/EMS/ems.json
ENV FIT_MAKE_TARGET=ci-skr

VOLUME /KUBECONFIG
VOLUME /EMS

RUN apk -U add build-base \
    curl \
    git \
    nodejs && \
    rm -rf /var/cache/apk/*

WORKDIR /
COPY clone-and-run-fit.sh /clone-and-run-fit.sh
RUN chmod +x /clone-and-run-fit.sh

CMD ["./clone-and-run-fit.sh"]
