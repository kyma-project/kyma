---
title: Installation with custom Istio deployment
type: Installation
---

You can use Kyma with a custom deployment of Istio that you installed in the target environment. To enable such implementation, remove Istio from the list of components that install with Kyma.
The version of your Istio deployment must match the version that Kyma currently supports.

In the installation process, the installer applies a custom patch to every Istio deployment. This is a mandatory step.  

>**NOTE:** To learn more, read the **Istio patch** document in the **Service Mesh** documentation topic.

## Prerequisites

- A live Istio version compatible with the version currently supported by Kyma. To check the supported version, see the value of the `REQUIRED_ISTIO_VERSION` environmental variable in the `resources/istio-kyma-patch/templates/job.yaml` file.
  >**NOTE:** Follow [this](https://istio.io/docs/setup/kubernetes/quick-start/) quick start guide to learn how to install and configure Istio on a Kubernetes cluster.

- Security enabled in your Istio deployment. To verify if security is enabled, check if the `policies.authentication.istio.io` custom resource exists in the cluster.
- Mutual TLS (mTLS) disabled in your Istio deployment.
- Kyma downloaded from the latest [release](https://github.com/kyma-project/kyma/releases).

## Local installation

1. Remove these lines from the `kyma-config-local.yaml` file:
  ```
  name: "istio"
  namespace: "istio-system"
  ```
2. Follow the installation steps described in the **Install Kyma locally from the release** document.

## Cluster installation

1. Remove these lines from the `kyma-config-cluster.yaml` file:
  ```
  name: "istio"
  namespace: "istio-system"
  ```
2. Follow the installation steps described in the **Install Kyma on a GKE cluster** document.

## Verify the installation

1. Check if all Pods are running in the `kyma-system` Namespace:
  ```
  kubectl get pods -n kyma-system
  ```
2. Sign in to the Kyma Console using the `admin@kyma.cx` as described in the **Install Kyma locally from the release** document.
