FROM lachlanevenson/k8s-kubectl:v1.10.6

COPY . /patch

RUN apk add bash

ENTRYPOINT /patch/patch.sh