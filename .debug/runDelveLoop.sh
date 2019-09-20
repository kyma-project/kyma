#!/bin/bash

# do horribleness to copy service account information to the spot that client-go expects it
#  TODO:  find a way to mount $TELEPRESENCE_ROOT as root.  proot?
K8S_SA_PATH=/var/run/secrets/kubernetes.io/serviceaccount

export KUBELESS_NAMESPACE="kyma-system"
export DOMAIN_NAME="34.93.71.163.xip.io"

mkdir -p $K8S_SA_PATH
cp $TELEPRESENCE_ROOT$K8S_SA_PATH/* $K8S_SA_PATH

while true; do
    echo Current workdir: $(pwd)
    ls

    echo "### looping dlv"
    {
        dlv debug --listen=0.0.0.0:2345 --headless=true --api-version 2

    } <&-

    echo "Enter valid path to a Go file"
    read filepath
    cd $filepath
done   
