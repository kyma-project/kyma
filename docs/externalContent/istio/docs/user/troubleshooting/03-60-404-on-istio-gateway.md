<!-- open-source-only -->
# You Get 404 Not Found 

## Symptom

When you try to access a Kyma endpoint, it reports a 404 error.

## Cause

The error might be caused by conflicts in the Istio Gateway host. For example, if you create two Gateways with the same host, Istio Ingress Gateway cannot reliably match an incoming request to a specific Gateway. As a result, requests receive 404 errors. Read the [Istio documentation](https://istio.io/latest/docs/ops/common-problems/network-issues/#404-errors-occur-when-multiple-gateways-configured-with-same-tls-certificate) to learn more about this behavior. 

Note that when you create `Ingress` resources using Istio as their ingress class, a `Gateway` entry is also created underneath. Use the `istioctl x internal-debug configz` command to check the cluster's configuration.

## Solution

Make sure that a host matches only one Gateway.
