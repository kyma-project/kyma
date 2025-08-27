<!-- open-source-only -->
# Expose a TCP Service Using Gateway API Alpha Support

This tutorial shows how to expose a TCP Service using Gateway API.

> [!WARNING]
> Exposing an unsecured workload to the outside world is a potential security vulnerability, so tread carefully. This tutorial is based on the experimental version of the Istio module, so it is not meant to be used in a production environment.

## Prerequisites

* The Istio module installation in the experimental version

## Steps

### Configure Gateway API Alpha Support

Edit the Istio custom resource by setting **enableAlphaGatewayAPI** to `true`:

```bash
kubectl patch istios/default -n kyma-system --type merge -p '{"spec":{"experimental":{"pilot": {"enableAlphaGatewayAPI": true}}}}'
```

### Install the experimental version of Gateway API CustomResourceDefinitions

The Istio module does not install Gateway API CustomResourceDefinitions (CRDs). To install the CRDs from the experimental channel, run the following command:

```bash
kubectl get crd gateways.gateway.networking.k8s.io &> /dev/null || \
{ kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd/experimental?ref=v1.1.0" | kubectl apply -f -; }
```

> [!NOTE]
> If you've already installed Gateway API CRDs from the standard channel, you must delete them before installing Gateway API CRDs from the experimental channel.

### Create a Workload

1. Export the name of the namespace in which you want to deploy the TCPEcho Service:

    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    ```

2. Create a namespace with Istio injection enabled and deploy the TCPEcho Service:

    ```bash
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    kubectl create -n $NAMESPACE -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/tcp-echo/tcp-echo.yaml
    ```

### Expose a TCPEcho Service

1. Create a Kubernetes Gateway to deploy Istio Ingress Gateway:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.networking.k8s.io/v1
    kind: Gateway
    metadata:
      name: tcp-echo-gateway
      namespace: ${NAMESPACE}
    spec:
      gatewayClassName: istio
      listeners:
      - name: tcp-31400
        port: 31400
        protocol: TCP
        allowedRoutes:
          namespaces:
            from: Same
    EOF
    ```

    > [!NOTE]
    > This command deploys the Istio Ingress service in your namespace with the corresponding Kubernetes Service of type `LoadBalanced` and an assigned external IP address.

2. Create a TCPRoute to configure access to your worklad:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.networking.k8s.io/v1alpha2
    kind: TCPRoute
    metadata:
      name: tcp-echo
      namespace: ${NAMESPACE}
    spec:
      parentRefs:
      - name: tcp-echo-gateway
        sectionName: tcp-31400
      rules:
      - backendRefs:
        - name: tcp-echo
          port: 9000
    EOF
    ```

### Send TCP Traffic to a TCPEcho Service

1. Discover Istio Ingress Gateway's IP and port:

    ```bash
    export INGRESS_HOST=$(kubectl get gtw tcp-echo-gateway -n $NAMESPACE -o jsonpath='{.status.addresses[0].value}')
    export INGRESS_PORT=$(kubectl get gtw tcp-echo-gateway -n $NAMESPACE -o jsonpath='{.spec.listeners[?(@.name=="tcp-31400")].port}')
    ```

2. Deploy a `sleep` Service:

    ```bash
    kubectl create -n $NAMESPACE -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/sleep/sleep.yaml
    ```


2. Send TCP traffic:

    ```bash
    export SLEEP=$(kubectl get pod -l app=sleep -n $NAMESPACE -o jsonpath={.items..metadata.name})
    for i in {1..3}; do \
    kubectl exec "$SLEEP" -c sleep -n $NAMESPACE -- sh -c "(date; sleep 1) | nc $INGRESS_HOST $INGRESS_PORT"; \
    done
    ```
    You should see similar output:
    ```
    hello Mon Jul 29 12:43:56 UTC 2024
    ```

