#!/bin/bash
set -o errexit

# Create docker network
docker network create kyma || echo "kyma network already exists"

# Start docker Registry
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name registry.localhost \
  --network kyma \
  -v $PWD/registry:/var/lib/registry \
  eu.gcr.io/kyma-project/test-infra/docker-registry-2:20200202

# Create Kyma cluster
k3d cluster create kyma \
    --image "docker.io/rancher/k3s:v1.19.7-k3s1" \
    --port 80:80@loadbalancer \
    --port 443:443@loadbalancer \
    --k3s-server-arg --no-deploy \
    --k3s-server-arg traefik \
    --network kyma \
    --volume $PWD/registries.yaml:/etc/rancher/k3s/registries.yaml \
    --wait \
    --kubeconfig-switch-context \
    --timeout 60s 

echo "Cluster created in $(( $SECONDS/60 )) min $(( $SECONDS % 60 )) sec"
