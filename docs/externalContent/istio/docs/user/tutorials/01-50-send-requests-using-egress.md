# Send Requests Using Istio Egress Gateway
Learn how to configure and use the Istio egress Gateway to allow outbound traffic from your Kyma runtime cluster to specific external destinations. Test your configuration by sending an HTTPS request to an external website using a sample Deployment.

## Prerequisites

* You have the Istio module added.
* To use CLI instruction, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
  and [curl](https://curl.se/).

## Steps

1. Export the following value as an environment variable:

    ```bash
    export NAMESPACE={service-namespace}
    ```

2. Create a new namespace for the sample application:
    ```bash
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

3. Enable the egress Gateway in the Istio custom resource:
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: operator.kyma-project.io/v1alpha2
   kind: Istio
   metadata:
     name: default
     namespace: kyma-system
     labels:
       app.kubernetes.io/name: default
   spec:
     components:
       egressGateway:
         enabled: true
   EOF
   ```

4. Enable additional sidecar logs to see the egress Gateway being used in requests:
    ```bash
    kubectl apply -f - <<EOF
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: mesh-default
      namespace: istio-system
    spec:
      accessLogging:
        - providers:
          - name: envoy
    EOF
    ```

5. Apply the `curl` Deployment to send the requests:
    ```bash
    kubectl apply -f - <<EOF
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: curl
      namespace: ${NAMESPACE}
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: curl
      namespace: ${NAMESPACE}
      labels:
        app: curl
        service: curl
    spec:
      ports:
      - port: 80
        name: http
      selector:
        app: curl
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: curl
      namespace: ${NAMESPACE}
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: curl
      template:
        metadata:
          labels:
            app: curl
        spec:
          terminationGracePeriodSeconds: 0
          serviceAccountName: curl
          containers:
          - name: curl
            image: curlimages/curl
            command: ["/bin/sleep", "infinity"]
            imagePullPolicy: IfNotPresent
            volumeMounts:
            - mountPath: /etc/curl/tls
              name: secret-volume
          volumes:
          - name: secret-volume
            secret:
              secretName: curl-secret
              optional: true
    EOF
    ```

6. Export the name of the `curl` Pod:
    ```bash
   export SOURCE_POD=$(kubectl get pod -n "$NAMESPACE" -l app=curl -o jsonpath={.items..metadata.name})
    ```

7. Define a ServiceEntry which adds the `kyma-project.io` hostname to the mesh:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: networking.istio.io/v1
   kind: ServiceEntry
   metadata:
     name: kyma-project
     namespace: $NAMESPACE
   spec:
     hosts:
     - kyma-project.io
     ports:
     - number: 443
       name: tls
       protocol: TLS
     resolution: DNS
   EOF
   ```

8. Create an egress Gateway, DestinationRule, and VirtualService to direct traffic:

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: networking.istio.io/v1
   kind: Gateway
   metadata:
     name: istio-egressgateway
     namespace: ${NAMESPACE}
   spec:
     selector:
       istio: egressgateway
     servers:
     - port:
         number: 443
         name: tls
         protocol: TLS
       hosts:
       - kyma-project.io
       tls:
         mode: PASSTHROUGH
   ---
   apiVersion: networking.istio.io/v1
   kind: DestinationRule
   metadata:
     name: egressgateway-for-kyma-project
     namespace: ${NAMESPACE}
   spec:
     host: istio-egressgateway.istio-system.svc.cluster.local
     subsets:
     - name: kyma-project
   ---
   apiVersion: networking.istio.io/v1
   kind: VirtualService
   metadata:
     name: direct-kyma-project-through-egress-gateway
     namespace: ${NAMESPACE}
   spec:
     hosts:
     - kyma-project.io
     gateways:
     - mesh
     - istio-egressgateway
     tls:
     - match:
       - gateways:
         - mesh
         port: 443
         sniHosts:
         - kyma-project.io
       route:
       - destination:
           host: istio-egressgateway.istio-system.svc.cluster.local
           subset: kyma-project
           port:
             number: 443
     - match:
       - gateways:
         - istio-egressgateway
         port: 443
         sniHosts:
         - kyma-project.io
       route:
       - destination:
           host: kyma-project.io
           port:
             number: 443
         weight: 100
   EOF
   ```

9. Send an HTTPS request to the Kyma project website:
   ```bash
   kubectl exec -n "$NAMESPACE" "$SOURCE_POD" -c curl -- curl -sSL -o /dev/null -D - https://kyma-project.io
   ```

   If successful, you get a response from the website similar to this one:
   ```
   HTTP/2 200
   accept-ranges: bytes
   age: 203
   ...
   ```

10. Check the logs of the Istio egress Gateway:
   ```bash
   kubectl logs -l istio=egressgateway -n istio-system
   ```

   If successful, the logs contain the request made by the egress Gateway:
   ```
   {"requested_server_name":"kyma-project.io","upstream_cluster":"outbound|443||kyma-project.io",[...]}
   ```

## Enhance Security by Implementing NetworkPolicies

Istio by default cannot securely enforce that egress traffic goes through the egress gateway. It only enables the flow
through sidecar proxies.
Kubernetes Network Policies can be used to restrict namespace traffic, so it only goes through the istio egress gateway.
Network policies are the Kubernetes way to enforce traffic rules in the namespace.

> [!NOTE] Support for NetworkPolicies depends on the Kubernetes CNI plugin used in the cluster. By default, SAP BTP, Kyma runtime uses the CNI configuration provided and managed by Gardener, which supports NetworkPolicies. However, if you’ve made any changes, make sure to check the relevant documentation.

> [!NOTE] In Gardener-based clusters, such as SAP BTP, Kyma runtime, the Network Policy restricting DNS traffic may not work as expected.
> It is due to the local DNS service used in discovery working outside the CNI. In such cases, define the **ipBlocks** with
> the IP CIDR of the `kube-dns` service in the NetworkPolicy to allow proper DNS resolution.

1. Fetch the IP address of the `kube-dns` service:
   ```bash
    export KUBE_DNS_ADDRESS=$(kubectl get svc -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}')
   ```
2. Create a NetworkPolicy with the fetched IP address in the **ipBlocks** section. The NetworkPolicy allows only egress traffic to the Istio egress gateway, blocking all other egress traffic.
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
      name: network-policy-allow-egress-traffic
      namespace: ${NAMESPACE}
   spec:
      egress:
      - ports:
        - port: 53
          protocol: UDP
        to:
        - namespaceSelector:
             matchLabels:
                kubernetes.io/metadata.name: kube-system
        - ipBlock:
             cidr: ${KUBE_DNS_ADDRESS}/32
      - to:
        - namespaceSelector:
             matchLabels:
                kubernetes.io/metadata.name: istio-system
      podSelector: {}
      policyTypes:
      - Egress
   EOF
   ```
3. Send an HTTPS request to the Kyma project website:
   ```bash
   kubectl exec -n "$NAMESPACE" "$SOURCE_POD" -c curl -- curl -sSL -o /dev/null -D - https://kyma-project.io
   ```
   If successful, you get a response from the website similar to this one:
   ```
   HTTP/2 200
   accept-ranges: bytes
   age: 203
   ...
   ```
4. Send an HTTPS request to an external website:
   ```bash
   kubectl exec -n "$NAMESPACE" "$SOURCE_POD" -c curl -- curl -sSL -o /dev/null -D - https://www.google.com
   ```
   The request should fail with an error message similar to this one:
   ```
   curl: (35) Recv failure: Connection reset by peer
   command terminated with exit code 35
   ```

You have successfully secured the egress traffic in the `$NAMESPACE` namespace by using Istio egress gateway and
Kubernetes Network Policies.
