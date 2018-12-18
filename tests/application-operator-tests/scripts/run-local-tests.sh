#!/usr/bin/env bash

CURRENT_DIR="$( cd "$(dirname "$0")" ; pwd -P )"

eval $(minikube docker-env)

echo ""
echo "------------------------"
echo "Removing test pod"
echo "------------------------"


kubectl -n kyma-integration delete po application-operator-tests --now

echo ""
echo "------------------------"
echo "Building tests image"
echo "------------------------"

docker build $CURRENT_DIR/.. -t application-operator-tests

echo ""
echo "------------------------"
echo "Creating test pod"
echo "------------------------"

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: application-operator-tests
spec:
  containers:
  - name: application-operator-tests
    image: application-operator-tests
    imagePullPolicy: Never
    env:
    - name: TILLER_HOST
      value: tiller-deploy.kube-system.svc.cluster.local:44134
    - name: NAMESPACE
      value: kyma-integration
    - name: INSTALLATION_TIMEOUT
      value: "180"
  restartPolicy: Never
EOF

echo ""
echo "------------------------"
echo "Waiting 5 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 5

kubectl -n kyma-integration logs application-operator-tests -f
