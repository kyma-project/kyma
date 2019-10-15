#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Building component image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t compass-runtime-agent

echo ""
echo "------------------------"
echo "Updating deployment"
echo "------------------------"

kubectl -n compass-system patch deployment compass-runtime-agent --patch 'spec:
  template:
    spec:
      containers:
      - name: compass-runtime-agent
        image: compass-runtime-agent
        imagePullPolicy: Never'

echo ""
echo "------------------------"
echo "Removing old pods"
echo "------------------------"

kubectl -n compass-system delete po -l app=compass-runtime-agent --now --wait=false
