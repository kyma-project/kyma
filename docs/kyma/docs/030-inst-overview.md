---
title: Overview
type: Installation
---

Kyma is a complex tool which consists of many different [components](#details-details) that provide various functionalities to extend your application. This entails high technical requirements that may influence your local development process. To meet the customers needs, we ensured Kyma modularity, which means that you can decide not to include a given component in the Kyma installation, or install it after the Kyma installation process.

To make the local development process easier, we introduced the **Kyma Lite** concept, in which case some components are by default not included in the local installation process. These are the Kyma and Kyma Lite components:

| Component | Kyma | Kyma Lite |
|----------------|------|------|
| `core` | ✅ | ✅ |
| `cluster-essentials` | ✅ | ✅ |
| `application-connector` | ✅ | ✅ |
| `ark` | ✅ | ⛔️ |
| `dex` | ✅ | ✅ |
| `istio` | ✅ | ✅ |
| `jaeger` | ✅ | ⛔️ |
| `logging` | ✅ | ⛔️ |
| `monitoring` | ✅ | ⛔️ |
| `prometheus-operator` | ✅ | ✅ |
| `service-catalog` | ✅ | ✅ |

### Installation guides

Follow these installation guides to install Kyma locally:

- [Install Kyma locally from release](#installation-install-kyma-locally-from-the-release)
- [Install Kyma locally from sources](#installation-install-kyma-locally-from-sources)

You can also install Kyma depending on the supported cloud providers:
- [Install Kyma on GKE cluster](#installation-install-kyma-on-a-gke-cluster)
- [Install Kyma on AKS cluster](#installation-install-kyma-on-an-aks-cluster)
- [Install Kyma on a GKE cluster with wildcard DNS](#installation-install-kyma-on-a-gke-cluster-with-wildcard-dns)
- [Install Kyma on a AKS cluster with wildcard DNS](#installation-install-kyma-on-an-aks-cluster-with-wildcard-dns)

Read rest of the installation documents to learn how to:
- [Install Kyma with Knative](#installation-installation-with-knative)
- [Disable a component's installation or install it later](#installation-custom-component-installation)
- [Update Kyma](#installation-update-kyma)
- [Reinstall Kyma](#installation-reinstall-kyma)
- [Get a deep-dive knowledge about the installation scripts](#installation-local-installation-scripts)

>**NOTE:** Make sure to check whether the version of the documentation in the left pane of the `kyma-project.io` is compatible with your Kyma version. 
