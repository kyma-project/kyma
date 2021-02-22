---
title: Overview
type: Installation
---

Kyma is a complex tool which consists of many different [components](#details-components) that provide various functionalities to extend your application. This entails high technical requirements that can influence your local development process. To meet the customer needs, we ensured Kyma modularity. This way you can decide not to include a given component in the Kyma deployment, or install it after the Kyma deployment process.

To make the local development process easier, we introduced the **Kyma Lite** concept in which case some components are not included in the local deployment process by default. These are the Kyma and Kyma Lite components in their deployment order:

| Component | Kyma | Kyma Lite |
|----------------|------|------|
| `cluster-essentials` | ✅ | ✅ |
| `istio` | ✅ | ✅ |
| `xip-patch` | ✅ | ✅ |
| `testing` | ✅ | ✅ |
| `logging` | ✅ | ⛔️ |
| `tracing` | ✅ | ⛔️ |
| `kiali` | ✅ | ⛔️ |
| `monitoring` | ✅ | ⛔️ |
| `knative-eventing` | ✅ | ✅ |
| `ory` | ✅ | ✅ |
| `api-gateway` | ✅ | ✅ |
| `service-catalog` | ✅ | ✅ |
| `service-catalog-addons` | ✅ | ✅ |
| `helm-broker` | ✅ | ✅ |
| `nats-streaming` | ✅ | ✅ |
| `core` | ✅ | ✅ |
| `cluster-users` | ✅ | ✅ |
| `serverless` | ✅ | ✅ |
| `knative-provisioner-natss` | ✅ | ✅ |
| `event-sources` | ✅ | ✅ |
| `application-connector` | ✅ | ✅ |
| `compass-runtime-agent` | ⛔️ | ⛔️ |

## Deployment guides

Follow these deployment guides to deploy Kyma locally or on a cluster:

- [Deploy Kyma locally](#installation-install-kyma-locally)
- [Deploy Kyma on a cluster](#installation-install-kyma-on-a-cluster)

Read the rest of the deployment documents to learn how to:
- [Disable the selected components' deployment or deploy them separately](#configuration-custom-component-installation)
- [Enable deployment profiles](#configuration-profiles)
- [Upgrade Kyma to a new version](#installation-upgrade-kyma)
- [Update Kyma](#installation-update-kyma)

>**NOTE:** Make sure that the version of the documentation selected in the left pane of `kyma-project.io` matches the version of Kyma you're using.
