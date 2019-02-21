ARG NODEIMAGE=kubeless/nodejs@sha256:5f1e999a1021dfb3d117106d80519a82110bd26a579f067f1ff7127025c90be5
FROM ${NODEIMAGE}

ADD start.sh /
USER root
RUN chmod +x /start.sh
USER 1000
ENTRYPOINT ["/start.sh"]