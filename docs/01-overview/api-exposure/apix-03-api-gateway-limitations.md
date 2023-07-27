---
title: API Gateway limitations
---

## Controller limitations

The APIRule controller is not a critical component of the Kyma networking infrastructure since it relies on Istio and Ory Custom Resources to provide routing capabilities. In terms of persistence, the controller depends on APIRules stored in the Kubernetes cluster. No other persistence solution is present.

In terms of the resource configuration, the following requirements are set on the API Gateway controller:

|          | CPU  | Memory |
|----------|------|--------|
| Limits   | 100m | 128Mi  |
| Requests | 10m  | 64Mi   |

## Limitations in terms of the number of created APIRules

The number of created APIRules is not limited. 

## Dependencies

API Gateway depends on Istio and Ory to provide its routing capabilities. In the case of the `allow` access strategy, only a Virtual Service Custom Resource is created. With any other access strategy, both Virtual Service and Oathkeeper Rule Custom Resource are created.
