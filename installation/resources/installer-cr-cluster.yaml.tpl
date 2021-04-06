apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: kyma-installation
  namespace: default
  labels:
    action: install
    kyma-project.io/installation: ""
spec:
  version: "__VERSION__"
  url: "__URL__"
  profile: ""
  components:
    - name: "cluster-essentials"
      namespace: "kyma-system"
    - name: "testing"
      namespace: "kyma-system"
    - name: "istio"
      namespace: "istio-system"
    - name: "xip-patch"
      namespace: "kyma-installer"
    - name: "dex"
      namespace: "kyma-system"
    - name: "ory"
      namespace: "kyma-system"
    - name: "api-gateway"
      namespace: "kyma-system"
    - name: "rafter"
      namespace: "kyma-system"
    - name: "service-catalog"
      namespace: "kyma-system"
    - name: "service-catalog-addons"
      namespace: "kyma-system"
    - name: "helm-broker"
      namespace: "kyma-system"
    - name: "eventing"
      namespace: kyma-system
    - name: "core"
      namespace: "kyma-system"
    - name: "cluster-users"
      namespace: "kyma-system"
    - name: "logging"
      namespace: "kyma-system"
    - name: "permission-controller"
      namespace: "kyma-system"
    - name: "apiserver-proxy"
      namespace: "kyma-system"
    - name: "iam-kubeconfig-service"
      namespace: "kyma-system"
    - name: "serverless"
      namespace: "kyma-system"
    - name: "application-connector"
      namespace: "kyma-integration"
    - name: "tracing"
      namespace: "kyma-system"
    - name: "monitoring"
      namespace: "kyma-system"
    - name: "kiali"
      namespace: "kyma-system"
    - name: "console"
      namespace: "kyma-system"

    #- name: "compass-runtime-agent"
    #  namespace: "compass-system"
