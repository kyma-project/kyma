# Istio Kyma patch

This chart packs the patch script as a Kubernetes job.

The Istio Kyma patch job runs and checks if the detected Istio deployment meets the following criteria:
  - A specific version of Istio is installed. The required version is defined in the [`values` file](https://github.com/kyma-project/kyma/blob/master/resources/istio-kyma-patch/values.yaml) of the patch.
  - Mutual TLS (mTLS) policy is enabled and set to `strict`.
  - [Istio policy enforcement](https://istio.io/docs/tasks/policy-enforcement/enabling-policy/) is enabled. 
  - Automatic sidecar injection is enabled.
  - Istio `policies.authentication.istio.io` CustomResourceDefinition (CRD) is present in the system.
