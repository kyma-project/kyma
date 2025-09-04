# Send mTLS Requests Using Istio Egress Gateway
Learn how to configure and use the Istio egress Gateway to allow mTLS-secured outbound traffic from your Kyma runtime cluster to a workload in another cluster.

## Prerequisites

* You need two clusters:
    1. One with Istio and API Gateway modules added. You will use it to expose the target workload.
    2. One with the Istio module added. You will use it to send requests.
* You must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl), [openssl](https://www.openssl.org), and [curl](https://curl.se/).

## Steps

### Generate mTLS Certificates

1. Export the `kubeconfig` file of the cluster that will contain the target workload.
    ```bash
    export KUBECONFIG={target-workload-cluster-config}
    ```

2. Export the following values as environment variables. You will use them throughout the whole tutorial.

    ```bash
    export DOMAIN={your-workload-host}{e.g. nginx.example.com}
    export CLIENT={client-cluster-domain}{e.g. client.example.com}
    export NAMESPACE={your-namespace}
    ```

3. Generate the root certificate:

    ```bash
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj "/O=example Inc./CN=$DOMAIN" -keyout egress.key -out egress.crt
    ```
    
4. Generate the host certificate:
    ```bash
    openssl req -out "$DOMAIN".csr -newkey rsa:2048 -nodes -keyout "$DOMAIN".key -subj "/CN="$DOMAIN"/O=some organization"
    openssl x509 -req -sha256 -days 365 -CA egress.crt -CAkey egress.key -set_serial 0 -in "$DOMAIN".csr -out "$DOMAIN".crt
    ```

5. Generate the client certificate:
    ```bash
    openssl req -out "$CLIENT".csr -newkey rsa:2048 -nodes -keyout "$CLIENT".key -subj "/CN="$CLIENT"/O=client organization"
    openssl x509 -req -sha256 -days 365 -CA egress.crt -CAkey egress.key -set_serial 1 -in "$CLIENT".csr -out "$CLIENT".crt
    ```

### Prepare a Cluster with a Workload

Use the same `kubeconfig` file you've already exported.

1. Create an HTTPBin Deployment:

    ```bash
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    kubectl -n $NAMESPACE create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
    ```

2. Create a Secret containing your generated certificates and key:

    ```bash
    kubectl create secret generic -n istio-system kyma-mtls-certs --from-file=cacert=egress.crt --from-file=key=$DOMAIN.key --from-file=cert=$DOMAIN.crt
    ```

3. Create a Gateway with mTLS configuration:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: networking.istio.io/v1alpha3
    kind: Gateway
    metadata:
     name: kyma-mtls-gateway
     namespace: $NAMESPACE
    spec:
     selector:
       app: istio-ingressgateway
       istio: ingressgateway
     servers:
       - port:
           number: 443
           name: mtls
           protocol: HTTPS
         tls:
           mode: MUTUAL
           credentialName: kyma-mtls-certs
         hosts:
           - $DOMAIN
    EOF
    ```

4. Create an APIRule exposing your workload:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v2alpha1
    kind: APIRule
    metadata:
     name: httpbin
     namespace: $NAMESPACE
    spec:
     hosts:
       - $DOMAIN
     service:
       name: httpbin
       namespace: $NAMESPACE
       port: 8000
     gateway: $NAMESPACE/kyma-mtls-gateway
     rules:
       - path: /*
         methods: ["GET"]
         noAuth: true
    EOF
    ```

5. Send a request to your workload without a certificate and key:

    ```bash
    curl -ik -X GET https://$DOMAIN/headers
    ```

    You should see an SSL error similar to this:
    `curl: (56) LibreSSL SSL_read: LibreSSL/3.3.6: error:1404C45C:SSL routines:ST_OK:reason(1116), errno 0`

6. Send a request with the certificate and key:

    ```bash
    curl --cert $CLIENT.crt --key $CLIENT.key --cacert egress.crt -ik -X GET https://$DOMAIN/headers
    ```

    You should see the `200` response code from the workload containing headers, one of them should be `"X-Forwarded-Client-Cert": ["xxx"]`

### Prepare a Cluster with an Egress Gateway

1. Export the `kubeconfig` file of another cluster:

    ```bash
    export KUBECONFIG={egress-cluster-config}
    ```

2. Create a new namespace for the sample application:

    ```bash
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

3. Create a Secret with the client certificate and key:

    ```bash
    kubectl create secret -n istio-system generic client-credential --from-file=tls.key=$CLIENT.key \
     --from-file=tls.crt=$CLIENT.crt --from-file=ca.crt=egress.crt
    ```

4. Enable the egress Gateway in the Istio custom resource:

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

5. Enable additional sidecar logs to see the egress Gateway being used in requests:

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

6. Apply the `curl` Deployment to send the requests:

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
      - port: 443
        name: https
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

7. Export the name of the `curl` Pod:
    ```bash
    export SOURCE_POD=$(kubectl get pod -n "$NAMESPACE" -l app=curl -o jsonpath={.items..metadata.name})
    ```

8. Define a ServiceEntry which adds the `kyma-project.io` hostname to the mesh:

    ```bash
    kubectl apply -f - <<EOF
    apiVersion: networking.istio.io/v1
    kind: ServiceEntry
    metadata:
      name: httpbin
      namespace: $NAMESPACE
    spec:
      hosts:
      - $DOMAIN
      ports:
      - number: 80
        name: http
        protocol: HTTP
      - number: 443
        name: https
        protocol: HTTPS
      resolution: DNS
    EOF
    ```

9. Create an egress Gateway, DestinationRule, and VirtualService to direct traffic:

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
         number: 80
         name: http
         protocol: HTTP
       hosts:
       - $DOMAIN
    ---
    apiVersion: networking.istio.io/v1
    kind: DestinationRule
    metadata:
     name: egressgateway-for-httpbin
     namespace: ${NAMESPACE}
    spec:
     host: istio-egressgateway.istio-system.svc.cluster.local
     subsets:
     - name: httpbin
    ---
    apiVersion: networking.istio.io/v1
    kind: VirtualService
    metadata:
     name: direct-httpbin-through-egress-gateway
     namespace: ${NAMESPACE}
    spec:
     hosts:
     - $DOMAIN
     gateways:
     - istio-egressgateway
     - mesh
     http:
     - match:
       - gateways:
         - mesh
         port: 80
       route:
       - destination:
           host: istio-egressgateway.istio-system.svc.cluster.local
           subset: httpbin
           port:
             number: 80
     - match:
       - gateways:
         - istio-egressgateway
         port: 80
       route:
       - destination:
           host: $DOMAIN
           port:
             number: 443
         weight: 100
    EOF
    ```

10. Create a DestinationRule to use mTLS credential:

    ```bash
    kubectl apply -n istio-system -f - <<EOF
    apiVersion: networking.istio.io/v1
    kind: DestinationRule
    metadata:
     name: originate-mtls-for-httpbin
    spec:
     host: $DOMAIN
     trafficPolicy:
       portLevelSettings:
       - port:
           number: 443
         tls:
           mode: MUTUAL
           credentialName: client-credential
           sni: $DOMAIN
    EOF
    ```

### Send HTTP Requests

1. Send an HTTP request to the Kyma project website:
    When you send an HTTP request, Istio uses egress to redirect the HTTPS to the website.

    ```bash
    kubectl exec -n $NAMESPACE "$SOURCE_POD" -c curl -- curl -ik -X GET http://$DOMAIN/headers
    ```

    If successful, you get the `200` response code from the workload containing headers. One of these headers should be `"X-Forwarded-Client-Cert": ["xxx"]`.

2. Check the logs of the Istio egress Gateway:
    ```bash
    kubectl logs -l istio=egressgateway -n istio-system
    ```

    If successful, the logs contain the request made by the egress Gateway:
    ```
    {"authority":"{YOUR_DOMAIN}":"outbound|443||{YOUR_DOMAIN}",[...]}
    ```

## Enhance Security by Implementing NetworkPolicies

By default, Istio cannot securely enforce that egress traffic is routed through the Istio egress gateway. It only enables the flow through sidecar proxies.
However, you can use Kubernetes NetworkPolicies to restrict namespace traffic, so it only passes through the Istio egress gateway.
NetworkPolicies are the Kubernetes method for enforcing traffic rules within a namespace.

> [!NOTE] Support for NetworkPolicies depends on the Kubernetes CNI plugin used in the cluster. By default, SAP BTP, Kyma runtime uses the CNI configuration provided and managed by Gardener, which supports NetworkPolicies. However, if you’ve made any changes, make sure to check the relevant documentation.

> [!NOTE] In Gardener-based clusters, such as SAP BTP, Kyma runtime, the Network Policy restricting DNS traffic may not work as expected.
> It is due to the local DNS service used in discovery working outside the CNI. In such cases, define the **ipBlocks** with
> the IP CIDR of the `kube-dns` service in the NetworkPolicy to allow proper DNS resolution.

1. Fetch the IP address of the `kube-dns` service:
   ```bash
    export KUBE_DNS_ADDRESS=$(kubectl get svc -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}')
   ```
2. Create a NetworkPolicy with the fetched IP address in the **ipBlocks** section. The NetworkPolicy allows only egress
   traffic to the Istio egress gateway, blocking all other egress traffic.
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
