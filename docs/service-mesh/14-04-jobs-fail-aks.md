---
title: Kubernetes jobs fail on AKS
type: Troubleshooting
---

A known issue related to Istio sidecar handling on AKS causes Kubernetes jobs with Istio Proxy sidecar to run endlessly as the sidecar doesn't terminate.
As a workaround, disable Istio sidecar injection for all jobs on AKS by adding the `sidecar.istio.io/inject: "false"` annotation.

To get a better understanding of this problem, read [this](https://github.com/istio/istio/issues/15041) Istio issue and the related discussion. 
