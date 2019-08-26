# Api-gateway

## Overview
API-Gateway is a component allowing exposure of a service thou the kyma Console. It deploys and manages Istio and Ory/Oathkeeper CRDs.

>**NOTE:** This CRD requires and uses the following applications/CRD, which should be installed beforehand:
> - Istio [VirtualService](https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/)
> - Istio [Policy](https://istio.io/docs/reference/config/istio.authentication.v1alpha1/)
> - Oathkeeper [AccessRule](https://www.ory.sh/docs/oathkeeper/)
>     + Oathkeeper CRD resources are available as charts in [this repo](https://github.com/ory/k8s)
