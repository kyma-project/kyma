#!/bin/bash
set -o errexit
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

echo "starting docker registry"
sudo mkdir -p /etc/rancher/k3s
sudo cp registries.yaml /etc/rancher/k3s
docker run -d \
-p 5000:5000 \
--restart=always \
--name registry.localhost \
-v $DIR/registry:/var/lib/registry \
eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202

echo "starting cluster"
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="v1.19.7+k3s1" K3S_KUBECONFIG_MODE=777 INSTALL_K3S_EXEC="server --disable traefik" sh -
mkdir -p ~/.kube
cp /etc/rancher/k3s/k3s.yaml ~/.kube/config
chmod 600 ~/.kube/config
