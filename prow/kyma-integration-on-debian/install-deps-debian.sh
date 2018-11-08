#!/bin/bash

###
# Following script installs necessary tooling for Debian to deploy Kyma on Minikube.
#
# REQUIREMENTS:
# 64-bit version of one of these Debian versions:
# 
# - Stretch 9 (stable)
# - Jessie 8 (LTS)
# 
###

set -o errexit

MINIKUBE_VERSION=v0.28.2
KUBECTL_CLI_VERSION=v1.10.0
HELM_VERSION=v2.8.2
DOCKER_VERSION=18.06.1~ce~3-0~debian

# install docker
sudo apt-get update
sudo apt-get install -y \
     apt-transport-https \
     ca-certificates \
     curl \
     gnupg2 \
     socat \
     software-properties-common

curl -fsSL https://download.docker.com/linux/debian/gpg | sudo apt-key add -

sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/debian \
   $(lsb_release -cs) \
   stable"

sudo apt update
sudo apt install -y docker-ce=${DOCKER_VERSION}

# install kubectl
curl -Lo /tmp/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_CLI_VERSION}/bin/linux/amd64/kubectl && \
 chmod +x /tmp/kubectl && \
 sudo mv /tmp/kubectl /usr/local/bin/kubectl

# install helm
curl -O https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
 tar -zxvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
 sudo mv linux-amd64/helm /usr/local/bin/helm && \
 rm -rf helm-${HELM_VERSION}-linux-amd64.tar.gz linux-amd64

# install minikube
curl -Lo /tmp/minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
 chmod +x /tmp/minikube && \
 sudo mv /tmp/minikube /usr/local/bin/minikube