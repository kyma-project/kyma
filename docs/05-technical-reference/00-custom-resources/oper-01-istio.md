---
title: Istio Manager Operator
---

The `istios.operator.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the Istio Manager Operator listens for. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```shell
kubectl get crd istios.operator.kyma-project.io -o yaml
```

More information on the Module Manager you may find [in here](https://github.com/kyma-project/module-manager/).

## Sample custom resource

This is a sample custom resource (CR) that the Istio Manager Operator listens for to manage Istio installation on the cluster. This example has only the **numTrustedProxies** configuration setting which makes the Istio Manager Operator install and configure Istio appropriately. There must be only a single Istio Manager CR in the cluster.

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: Istio
metadata:
  name: istio-manager
  labels:
    app.kubernetes.io/name: istio-manager
spec:
  config:
    numTrustedProxies: 1
```

This table lists all the possible parameters of a given resource together with their descriptions:

| Field   |      Mandatory      |  Description |
|---|:---:|---|
| **metadata.name** | **YES** | Specifies the name of the Istio Manager CR. |
| **spec.config.numTrustedProxies** | **NO** | Specifies the number of trusted proxies, [more info here](https://istio.io/latest/docs/ops/configuration/traffic-management/network-topologies/). |

## Additional information

When you fetch the existing Istio Manager CR, there is the **status** section which describes the status of the Istio installation on the cluster. This table lists the fields of the **status** section.

| Field   |  Description |
|:---|:---|
| **status.state** | State describing the Istio installation status. |
| **status.conditions** | Addition details as conditions. |

### Status states

These are the status states used to describe the Istio installation:

| Code   |  Description |
|---|---|
| **Ready** | Istio was installed successfully. |
| **Processing** | Istio installation currently in progress. |
| **Error** | Istio installation failed. |
| **Deleting** | Istio is being removed. |
