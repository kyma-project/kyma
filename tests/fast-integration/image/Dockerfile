FROM alpine:3.14.6

ENV KUBECONFIG=/KUBECONFIG/kubeconfig.yaml
ENV EVENTMESH_SECRET_FILE=/EMS/ems.json
ENV FIT_MAKE_TARGET=ci-skr

VOLUME /KUBECONFIG
VOLUME /EMS

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

RUN apk --no-cache add \
    build-base \
    git \
    nodejs=14.19.0-r0

WORKDIR /
COPY clone-and-run-fit.sh /clone-and-run-fit.sh
RUN chmod +x /clone-and-run-fit.sh

CMD ["./clone-and-run-fit.sh"]
