---
title: Install subcomponents
type: Details
---

It is up to you to decide which subcomponents you install as part of the `core` release. By default, most of the core subcomponents are enabled. If you want to install only specific subcomponents, follow the steps that you need to perform before the local and cluster installation.

## Install subcomponents locally

To specify whether to install a given core subcomponent on Minikube, use the `manage-component.sh` script before you trigger the Kyma installation. The script consumes two parameters:

- the name of the core subcomponent
- a Boolean value that determines whether to install the subcomponent (`true`) or not (`false`)

Example:

To enable the `Azure Broker` subcomponent, run the following command:
```
scripts/manage-component.sh azure-broker true
```

Alternatively, to disable the `Azure Broker` subcomponent, run this command:
```
scripts/manage-component.sh azure-broker false
```

## Install subcomponents on a cluster

Install subcomponents on a cluster based on Helm conditions described in the `requirements.yaml` file. Read more about the fields in the `requirements.yaml` file [here](https://github.com/helm/helm/blob/master/docs/charts.md#tags-and-condition-fields-in-requirementsyaml).

To specify whether to install a given core subcomponent, provide override values before you trigger the installation.

Example:
```
apiVersion: v1
kind: ConfigMap
metadata:
  name: kyma-sub-components
  namespace: kyma-installer
  labels:
    installer: overrides
data:
  azure-broker.enabled: "true"
```

>**NOTE:** Some subcomponents can require additional configuration to work properly.