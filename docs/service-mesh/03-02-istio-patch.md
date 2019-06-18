---
title: Istio patch
type: Details
---

As a core component, Istio installs with every Kyma deployment by default. The installation consists of two steps:

1. Istio installs using the official, charts from the currently supported release. The charts that are currently
used are stored in the `resources/istio` directory. The installation is customized as described in [this document](#details-istio-customization)

2. A Istio patch is run, which verifies if the current installation has the following options:
  - A specific version is Istio is installed
  - mTLS(Mutual TLS) policy is enabled and set to `strict`
  - Istio policy enforcement is enabled 
  - Automatic sidecar injection is enabled
  - The Istio CRD(CustomResourceDefinition) `policies.authentication.istio.io` is present in the system

If any of the above options is missing, the patch fails, which results in a failed installation. In such a case logs of the patch can give some insight into which options are missing:

```
    kubectl logs -n istio-system -l app=istio-kyma-patch
```

To learn more about the custom Istio patch applied in Kyma, see the `components/istio-kyma-patch/` directory.

>**NOTE:** This verification patch is an optional component and can be disabled. However, it is enabled by default

## Use an existing Istio installation with Kyma

You can use an existing installation of Istio running on Kubernetes with Kyma. In such a case is is recommended to enable the patch component to verify if all the required options are set. 

To allow such implementation, you must install Kyma without Istio. Read [this](/root/kyma#configuration-custom-component-installation) document for more details.
