#!/usr/bin/env bash

./tests/components/application-connector/scripts/jobguard.sh

service docker start
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
curl -Lo kyma.tar.gz "https://github.com/kyma-project/cli/releases/download/$(curl -s https://api.github.com/repos/kyma-project/cli/releases/latest | grep tag_name | cut -d '"' -f 4)/kyma_Linux_x86_64.tar.gz" && mkdir kyma-release && tar -C kyma-release -zxvf kyma.tar.gz && chmod +x kyma-release/kyma && rm -rf kyma.tar.gz
k3d cluster create
kubectl create ns kyma-system
kubectl cluster-info
kyma-release/kyma deploy --ci --components-file tests/components/application-connector/resources/installation-config/mini-kyma-skr.yaml --source local --workspace $PWD
cd tests/components/application-connector

# reconfigure DNS
kubectl apply -f resources/patches/coredns.yaml
kubectl -n kube-system delete pods -l k8s-app=kube-dns

make -f Makefile.test-compass-runtime-agent test-compass-runtime-agent
failed=$?

k3d cluster delete kyma
exit $failed
