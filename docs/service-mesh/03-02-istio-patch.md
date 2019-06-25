---
title: Istio patch
type: Details
---

As a core component, Istio installs with every Kyma deployment by default. The installation consists of two steps:

1. Istio installs using the official charts from the currently supported release. The charts that are currently used are stored in the `resources/istio` directory. The installation is customized as described in [this document](#details-istio-customization)

2. A custom Istio patch job runs and checks if the detected Istio deployment meets the following criteria:
  - A specific version of Istio is installed. The required version is defined in the [`values` file](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/values.yaml#L11) of the patch.
  - Mutual TLS (mTLS) policy is enabled and set to `strict`.
  - [Istio policy enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/) is enabled. 
  - Automatic sidecar injection is enabled.
  - Istio `policies.authentication.istio.io` CustomResourceDefinition (CRD) is present in the system.

If the Istio deployment fails to meet any of these criteria, the patch fails, which results in a failed installation. If the installation failed, get the Istio patch logs to find out which criteria the Istio deployment didnt' meet. Run: 

```
kubectl logs -n istio-system -l app=istio-kyma-patch
```

To learn more about the custom Istio patch applied in Kyma, see the `components/istio-kyma-patch/` directory.

>**NOTE:** Istio patch is an optional component and can be disabled. However, it is enabled by default. Read [this](/root/kyma/#configuration-custom-component-installation) to learn how to disable components. 

## Use an existing Istio installation with Kyma

You can use an existing installation of Istio running on Kubernetes with Kyma. In such a case is is required to enable the patch component to verify if all the required options are set. 

To allow such implementation, you must install Kyma without Istio. Read [this](/root/kyma/#configuration-custom-component-installation) document for more details.
