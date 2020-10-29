---
title: Certificates chart
type: Configuration
---

The Certificates chart contains logic related to TLS certificate handling in Kyma.
It is important to install it in the same namespace that Istio is installed in (istio-system), in order to ensure TLS-related secrets are visible to Istio.
The chart consists of four sub-modules (subcharts), each one handling a specific mode: `gardener`, `xip`, `legacy` and `user-provided`.
Please refer to the description of available modes in [certificate management](docs/security/../../03-06-certificate-management.md) for details about modes configuration.
The chart does not expose any user-configurable values.
