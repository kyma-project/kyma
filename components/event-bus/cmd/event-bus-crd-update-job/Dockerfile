FROM eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20190325-ff66a3a
LABEL source=git@github.com:kyma-project/kyma.git

WORKDIR /root/
ADD update-crds.sh .
ADD crds.yaml .

ENTRYPOINT [ "/root/update-crds.sh" ]