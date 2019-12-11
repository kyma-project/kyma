---
title: Overview
type: Installation
---

Kyma is a complex tool which consists of many different [components](#details-components) that provide various functionalities to extend your application. This entails high technical requirements that can influence your local development process. To meet the customer needs, we ensured Kyma modularity. This way you can decide not to include a given component in the Kyma installation, or install it after the Kyma installation process.

To make the local development process easier, we introduced the **Kyma Lite** concept in which case some components are not included in the local installation process by default. These are the Kyma and Kyma Lite components:

| Component | Kyma | Kyma Lite |
|----------------|------|------|
| `core` | ✅ | ✅ |
| `cms` | ✅ | ✅ |
| `cluster-essentials` | ✅ | ✅ |
| `application-connector` | ✅ | ✅ |
| `backup-init` | ✅ | ⛔️ |
| `backup`  | ✅ | ⛔️ |
| `assetstore` | ✅ | ✅ |
| `dex` | ✅ | ✅ |
| `helm-broker` | ✅ | ✅ |
| `istio` | ✅ | ✅ |
| `istio-kyma-patch` | ✅ | ✅ |
| `jaeger` | ✅ | ⛔️ |
| `logging` | ✅ | ⛔️ |
| `monitoring` | ✅ | ⛔️ |
| `kiali` | ✅ | ⛔️ |
| `service-catalog` | ✅ | ✅ |
| `service-catalog-addons` | ✅ | ✅ |
| `nats-streaming` | ✅ | ✅ |
| `ory` | ✅ | ✅ |
| `knative-serving` | ✅ | ✅ |
| `knative-eventing` | ✅ | ✅ |
| `knative-build` | ⛔ | ⛔ |

>**NOTE**: To include backup in the installation process, you need to set it up first. Use the Kyma backup [setup instructions](/components/backup#overview-overview) to do so.

## Installation guides

Follow these installation guides to install Kyma locally or on a cluster:

- [Install Kyma locally](#installation-install-kyma-locally)
- [Install Kyma on a cluster](#installation-install-kyma-on-a-cluster)

Read the rest of the installation documents to learn how to:
- [Disable the selected components' installation or install them separately](#configuration-custom-component-installation)
- [Upgrade Kyma to a new version](#installation-upgrade-kyma)
- [Update Kyma](#installation-update-kyma)
- [Reinstall Kyma](#installation-reinstall-kyma)
- [Set up the Kyma backup](/components/backup#overview-overview)

>**NOTE:** Make sure that the version of the documentation selected in the left pane of `kyma-project.io` matches the version of Kyma you're using.
