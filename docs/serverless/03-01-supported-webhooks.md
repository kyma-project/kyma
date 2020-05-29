---
title: Supported webhooks
type: Details
---

A newly created or updated Function CR is verified by these two webhooks:

1. **Defaulting webhook** that sets the optimal [default values](https://github.com/kyma-project/kyma/blob/master/components/function-controller/pkg/apis/serverless/v1alpha1/function_defaults.go#L15) for CPU and memory requests and limits, and adds the maximum and the minimum number of replicas, if not specified already in the Function CR.

2. **Validation webhook** that checks if:

- Minimum values requested for CPU, memory, and replicas are not lower than the [required ones](https://github.com/kyma-project/kyma/blob/master/components/function-controller/pkg/apis/serverless/v1alpha1/function_validation.go#L19).
- Requests are lower than or equal to limits, same for replicas.
- The Function CR contains all the required parameters.
- The format of deps, envs, labels, and the Function name ([RFC 1123](https://tools.ietf.org/html/rfc1123)) is correct.
- The Function CR contains any [envs reserved for the KService](https://github.com/kyma-project/kyma/blob/master/resources/serverless/charts/webhook/values.yaml#L61).
