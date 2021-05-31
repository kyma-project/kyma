---
title: Prepare a cluster
type: Getting Started
---

first doc

## Profiles

By default, Kyma is installed with the default chart values defined in the `values.yaml` files. However, you can also install Kyma with the pre-defined profiles that differ in the amount of resources, such as memory and CPU, that the components can consume. The currently supported profiles are:
- Evaluation - a profile with limited resources that you can use for trial purposes
- Production - a profile configured for high availability and scalability. It requires more resources than the evaluation profile but is a better choice for production workload.

You can check the values used for each component in respective folders of the [`resources`](https://github.com/kyma-project/kyma/tree/master/resources) directory. The `profile-evaluation.yaml` file contains values used for the evaluation profile, and the `profile-production.yaml` file contains values for the production profile. If the component doesn't have files for respective profiles, the profile values are the same as default chart values defined in the `values.yaml` file.

A profile is defined globally for the whole Kyma installation. It's not possible to install a profile only for the selected components. However, you can set [overrides](#configuration-helm-overrides-for-kyma-installation) to override values set for the profile. The profile values have precedence over the default chart values, and overrides have precedence over the applied profile.

To install Kyma with any of the predefined profiles, follow the instructions described in the [cluster Kyma installation](#installation-install-kyma-on-a-cluster) document and set a profile with the `--profile` flag, as described in the [Install Kyma](#installation-install-kyma-on-a-cluster-install-kyma) section.

>**NOTE:** You can also set profiles on a running cluster during the [Kyma upgrade operation](#installation-upgrade-kyma).
