FROM curlimages/curl:7.66.0
LABEL source = git@github.com:kyma-project/kyma.git
COPY before-backup.test /
COPY assets /assets
COPY licenses /licenses
ENTRYPOINT ["/before-backup.test"]
