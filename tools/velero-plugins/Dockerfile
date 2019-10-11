FROM alpine:3.8
RUN mkdir /plugins
ADD velero-plugins /plugins/
USER nobody:nobody
ENTRYPOINT ["/bin/ash", "-c", "cp /plugins/* /target/."]
