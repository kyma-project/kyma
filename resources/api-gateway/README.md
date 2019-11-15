# API-Gateway

## Overview
API-Gateway is a component that allows exposing services through the kyma Console. It deploys and manages Istio and Ory/Oathkeeper Custom Resource Definitions (CRDs).

This chart installs the controller, which requires these CRDs to expose services:
- Istio [VirtualService](https://istio.io/docs/reference/config/networking/virtual-service/)
- Istio [Policy](https://istio.io/docs/reference/config/security/istio.authentication.v1alpha1/)
- Oathkeeper [AccessRule](https://www.ory.sh/docs/oathkeeper/)

>**NOTE:** Oathkeeper CRD resources are available as charts in [this](https://github.com/ory/k8s) repository.
