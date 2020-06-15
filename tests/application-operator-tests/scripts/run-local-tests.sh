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
  annotations:
   sidecar.istio.io/inject: "false"
spec:
  containers:
  - name: application-operator-tests
    image: application-operator-tests
    imagePullPolicy: Never
    env:
    - name: NAMESPACE
      value: kyma-integration
    - name: INSTALLATION_TIMEOUT_SECONDS
      value: "180"
    - name: HELM_DRIVER
      value: secret
  restartPolicy: Never
EOF

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    helm-chart-test: "true"
  name: app-operator-tests-role
  namespace: kyma-integration
rules:
- apiGroups:
  - '*'
  resources:
  - pods/log
  verbs:
  - get
  - list
- apiGroups:
  - 'networking.istio.io'
  resources:
  - virtualservices
  verbs:
  - get
  - list
EOF

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    helm-chart-test: "true"
  name: app-operator-tests-clusterrole
rules:
- apiGroups:
  - 'rbac.authorization.k8s.io'
  resources:
  - clusterroles
  - clusterrolebindings
  verbs:
  - get
  - list
- apiGroups:
  - '*'
  resources:
  - roles
  - rolebindings
  - services
  - serviceaccounts
  - deployments
  verbs:
  - get
  - list
- apiGroups:
  - '*'
  resources:
  - namespaces
  - configmaps
  - secrets
  verbs:
  - get
  - list
  - update
  - create
  - delete
- apiGroups:
  - '*'
  resources:
  - pods
  verbs:
  - get
  - list
  - update
  - create
  - delete
  - watch
- apiGroups:
  - 'servicecatalog.k8s.io'
  resources:
  - serviceinstances
  verbs:
  - get
  - list
  - update
  - create
  - delete
EOF

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    helm-chart-test: "true"
  name: app-operator-tests-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: app-operator-tests-clusterrole
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kyma-integration:default
EOF

cat <<EOF | kubectl -n kyma-integration apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    helm-chart-test: "true"
  name: app-operator-tests-rolebinding
  namespace: kyma-integration
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: app-operator-tests-role
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:serviceaccount:kyma-integration:default
EOF

echo ""
echo "------------------------"
echo "Waiting 5 seconds for pod to start..."
echo "------------------------"
echo ""

sleep 10

kubectl -n kyma-integration logs application-operator-tests -f
