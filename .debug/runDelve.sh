#!/bin/bash

# do horribleness to copy service account information to the spot that client-go expects it
#  TODO:  find a way to mount $TELEPRESENCE_ROOT as root.  proot?
K8S_SA_PATH=/var/run/secrets/kubernetes.io/serviceaccount

mkdir -p $K8S_SA_PATH
cp $TELEPRESENCE_ROOT$K8S_SA_PATH/* $K8S_SA_PATH

# run delve in a loop or just once
echo "### single dlv"
dlv debug --listen=0.0.0.0:2345 --headless=true --api-version 2
