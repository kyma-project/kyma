---
title: Overview
type: Installation
---

Kyma is a complex tool which consists of many different [components](#details-components) that provide various functionalities to extend your application. This entails high technical requirements that can influence your local development process. To meet the customer needs, we ensured Kyma modularity. This way you can decide not to include a given component in the Kyma installation, or install it after the Kyma installation process.

To make the local development process easier, we introduced the **Kyma Lite** concept in which case some components are not included in the local installation process by default. These are the Kyma and Kyma Lite components in their installation order:

| Component | Kyma | Kyma Lite |
|----------------|------|------|
| `cluster-essentials` | ✅ | ✅ |
| `testing` | ✅ | ✅ |
| `istio` | ✅ | ✅ |
| `xip-patch` | ✅ | ✅ |
| `knative-eventing` | ✅ | ✅ |
| `knative-eventing-kafka` | ⛔️ | ⛔️ |
| `dex` | ✅ | ✅ |
| `ory` | ✅ | ✅ |
| `api-gateway` | ✅ | ✅ |
| `rafter` | ✅ | ✅ |
| `service-catalog` | ✅ | ✅ |
| `service-catalog-addons` | ✅ | ✅ |
| `helm-broker` | ✅ | ✅ |
| `nats-streaming` | ✅ | ✅ |
| `core` | ✅ | ✅ |
| `cluster-users` | ✅ | ✅ |
| `permission-controller` | ✅ | ✅ |
| `apiserver-proxy` | ✅ | ✅ |
| `iam-kubeconfig-service` | ✅ | ✅ |
| `serverless` | ✅ | ✅ |
| `knative-provisioner-natss` | ✅ | ✅ |
| `event-sources` | ✅ | ✅ |
| `application-connector-ingress` | ✅ | ✅ |
| `application-connector-helper` | ✅ | ✅ |
| `application-connector` | ✅ | ✅ |
| `logging` | ✅ | ⛔️ |
| `tracing` | ✅ | ⛔️ |
| `monitoring` | ✅ | ⛔️ |
| `kiali` | ✅ | ⛔️ |
| `compass-runtime-agent` | ⛔️ | ⛔️ |

## Profiles

By default, Kyma is installed with the default chart values defined in the `values.yaml` files. However, you can also install Kyma with the pre-defined profiles that differ in the amount of resources, such as memory and CPU, that the components can consume. The currently supported profiles are:
- Evaluation - a profile with limited resources that you can use for trial purposes
- Production - a profile configured for high availability and scalability. It requires more resources than the evaluation profile but is a better choice for production workload.

You can check the values used for each component in respective folders of the [`resources`](https://github.com/kyma-project/kyma/tree/master/resources) directory. The `profile-evaluation.yaml` file contains values used for the evaluation profile, and the `profile-production.yaml` file contains values for the production profile. If the component doesn't have files for respective profiles, the profile values are the same as default chart values defined in the `values.yaml` file.

A profile is defined globally for the whole Kyma installation. It's not possible to install a profile only for the selected components. However, you can set [overrides](#configuration-helm-overrides-for-kyma-installation) to override values set for the profile. The profile values have precedence over the default chart values, and overrides have precedence over the applied profile.

To install Kyma with any of the predefined profiles, follow the instructions described in the [cluster Kyma installation](#installation-install-kyma-on-a-cluster) document and set a profile with the `--profile` flag, as described in the [Install Kyma](#installation-install-kyma-on-a-cluster-install-kyma) section.

>**NOTE:** You can also set profiles on a running cluster during the [Kyma upgrade operation](#installation-upgrade-kyma).

## Installation guides

Follow these installation guides to install Kyma locally or on a cluster:

- [Install Kyma locally](#installation-install-kyma-locally)
- [Install Kyma on a cluster](#installation-install-kyma-on-a-cluster)

Read the rest of the installation documents to learn how to:
- [Disable the selected components' installation or install them separately](#configuration-custom-component-installation)
- [Enable installation profiles](#configuration-profiles)
- [Upgrade Kyma to a new version](#installation-upgrade-kyma)
- [Update Kyma](#installation-update-kyma)

>**NOTE:** Make sure that the version of the documentation selected in the left pane of `kyma-project.io` matches the version of Kyma you're using.
