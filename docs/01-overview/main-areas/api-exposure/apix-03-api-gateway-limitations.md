---
title: API Gateway limitations
---

## Controller limitations

The API Rule controller is not a critical component of Kyma networking infrastructure, since it relies on Istio and Ory Custom Resources to provide routing capabilities. In terms of persistence the controller depends on API Rules stored in the Kubernetes cluster and no other persistence solution is present.

In terms of resource configuration set on API Gateway controller they are as follows:

|          | CPU  | Memory |
|----------|------|--------|
| Limits   | 100m | 128Mi  |
| Requests | 10m  | 64Mi   |

## Limitations in terms of number of created API Rules

There is no limitations in terms of number of created API Rules.

## Dependencies

API Gateway depends on Istio and Ory for providing its routing capabilities. In case of access strategy `allow` only a Virtual Service Custom Resource is created and with any other a Virtual Service and Oathkeeper Rule Custom Resource will be created.
