---
title: Istio patch
type: Details
---

As a core component, Istio is installed with every Kyma Deployment by default. To run Kyma, this default Istio installation needs some changes. The [Istio Kyma patch](../../components/istio-kyma-patch/README.md) application patches the already existing Istio installation so that Kyma can use it. Considering the above, the installation includes the following steps:

1. Istio is installed using the official charts from the currently supported release. The charts reside in the `resources/istio` directory.

2. A custom Istio patch job runs and checks if the detected Istio Deployment meets the following criteria:
  - A specific version of Istio is installed. The required version resides in the [`values`](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/values.yaml#L11) file.
  - Mutual TLS (mTLS) policy is enabled and set to `strict`.
  - [Istio policy enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/) is enabled. 
  - Automatic sidecar injection is enabled.
  - Istio `policies.authentication.istio.io` CustomResourceDefinition (CRD) is present in the system.

If the Istio Deployment fails to meet any of these criteria, the patch fails, which results in a failed installation. In such a case, get the Istio patch logs to find out which criteria the Istio Deployment didn't meet. Run: 

```bash
kubectl logs -n istio-system -l app=istio-kyma-patch
```

>**NOTE:** The Istio patch component is enabled by default, but you can disable it at any time using [these](/root/kyma/#configuration-custom-component-installation) instructions. 

## Use an existing Istio installation with Kyma

You can use an existing installation of Istio with your Kyma installation. Perform the following steps:

1. Make sure the Istio Patch component is enabled and meets the criteria required by the Istio patch job.
2. Install Kyma without Istio. Read [this](/root/kyma/#configuration-custom-component-installation) document for more details.
