# Dependencies of the API Gateway Module

## Istio Dependency

To use API Gateway, you must install Istio on your cluster. This is required because API Gateway creates the custom resources `Gateway` and `VirtualService`, which are provided by Istio. The recommended method for installing Istio is by using [Kyma Istio Operator](https://github.com/kyma-project/istio#install-kyma-istio-operator-and-istio-from-the-latest-release).

## Dependencies in SAP BTP, Kyma Runtime

In SAP BTP, Kyma runtime, API Gateway uses `DNSEntry` and `Certificate` custom resources provided by [Gardener](https://gardener.cloud). Because SAP BTP, Kyma runtime uses Gardener as the running environment, managed offering users are not required to install any dependencies manually.