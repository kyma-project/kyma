---
title: Overview
type: Installation
---

Kyma is a complex tool which consists of many different [components](#details-components) that provide various functionalities to extend your application. This entails high technical requirements that can influence your local development process. To meet the customer needs, we ensured Kyma modularity. This way you can decide not to include a given component in the Kyma deployment, or install it after the Kyma deployment process.

These are the Kyma prerequisite components, in their deployment order:

- `cluster-essentials`
- `istio`
- `certificates`

These are the Kyma regular components, deployed in parallel in random order:

- `testing`
- `logging`
- `tracing`
- `kiali`
- `monitoring`
- `eventing`
- `ory`
- `api-gateway`
- `service-catalog`
- `service-catalog-addons`
- `rafter`
- `helm-broker`
- `cluster-users`
- `serverless`
- `application-connector`

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
