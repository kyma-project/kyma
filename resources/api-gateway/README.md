# API-Gateway

## Overview
API-Gateway is a component that allows exposing services through the kyma Console. It deploys and manages Istio and Ory/Oathkeeper custom resource definitions (CRDs).

This chart installs the controller, which requires these CRDs to expose services:
- Istio [Virtual Service](https://istio.io/docs/reference/config/networking/virtual-service/)
- Oathkeeper [Rule](https://www.ory.sh/docs/oathkeeper/)

>**NOTE:** Oathkeeper CRD resources are available as charts in [this](https://github.com/ory/k8s) repository.
