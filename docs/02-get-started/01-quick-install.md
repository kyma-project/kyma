# Quick Install

To get started with Kyma, let's quickly install it with specific modules first.

> **NOTE:** This guide describes installation of standalone Kyma with specific modules. If you are using SAP BTP, Kyma runtime (SKR), read [Enable and Disable a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud) instead.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes cluster, or [k3d](https://k3d.io) (v5.x or higher) for local installation
- `kyma-system` namespace created

## Steps

1. Provision a k3d cluster, run:

  ```bash
  kyma provision k3d
  ```

2. Choose a module, deploy its module manager, and apply the module configuration. The operation installs a Kyma module of your choice on a Kubernetes cluster. See the already available Kyma modules with their quick installation steps and links to their GitHub repositories:

  [**Application Connector**](https://github.com/kyma-project/application-connector-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/application-connector-manager/releases/latest/download/application-connector-manager.yaml
  kubectl apply -f https://github.com/kyma-project/application-connector-manager/releases/latest/download/default_application_connector_cr.yaml -n kyma-system
  ```

  [**Keda**](https://github.com/kyma-project/keda-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
  kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml -n kyma-system
  ```

  [**SAP BTP Operator**](https://github.com/kyma-project/btp-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-manager.yaml
  kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator-default-cr.yaml -n kyma-system
  ```

  > **CAUTION:** The CR is in the `Warning` state and the message is `Secret resource not found reason: MissingSecret`. To create a Secret, follow the instructions in the [`btp-manager`](https://github.com/kyma-project/btp-manager/blob/main/docs/user/02-10-usage.md#create-and-install-secret) repository.

  [**Serverless**](https://github.com/kyma-project/serverless-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
  kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml  -n kyma-system
  ```

  [**Telemetry**](https://github.com/kyma-project/telemetry-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/telemetry-manager.yaml
  kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/telemetry-default-cr.yaml -n kyma-system
  ```

  [**NATS**](https://github.com/kyma-project/nats-manager)

  ```bash
  kubectl apply -f https://github.com/kyma-project/nats-manager/releases/latest/download/nats-manager.yaml
  kubectl apply -f https://github.com/kyma-project/nats-manager/releases/latest/download/nats_default_cr.yaml -n kyma-system
  ```

  [**API Gateway**](https://github.com/kyma-project/api-gateway)

  ```bash
  kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/api-gateway-manager.yaml
  kubectl apply -f https://github.com/kyma-project/api-gateway/releases/latest/download/apigateway-default-cr.yaml
  ```
  
  [**Istio**](https://github.com/kyma-project/istio)
  ```bash
  kubectl label namespace kyma-system istio-injection=enabled --overwrite
  kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
  kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
  ```

  [**Istio**](https://github.com/kyma-project/istio)
  ```bash
  kubectl label namespace kyma-system istio-injection=enabled --overwrite
  kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-manager.yaml
  kubectl apply -f https://github.com/kyma-project/istio/releases/latest/download/istio-default-cr.yaml
  ```

3. To manage Kyma using graphical user interface (GUI), open Kyma Dashboard:

  ```bash
  kyma dashboard
  ```
  <!-- markdown-link-check-disable-next-line -->
  This command takes you to your Kyma dashboard under [`http://localhost:3001/`](http://localhost:3001/).

## Related Links

- To see the list of all available Kyma modules, go to [Kyma modules](../06-modules/README.md).
- To learn how to [uninstall and upgrade Kyma with a module](./08-uninstall-upgrade-kyma-module.md).
