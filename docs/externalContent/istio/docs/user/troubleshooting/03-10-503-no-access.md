<!-- open-source-only -->
# Can't Access a Kyma Endpoint (503 status code)

## Symptom

You try to access a Kyma endpoint and receive the `503` status code.

## Cause

This behavior might be caused by a configuration error in Istio Ingress Gateway. As a result, the endpoint you call is not exposed.

## Solution

To fix this problem, restart the Pods of Istio Ingress Gateway.

<Tabs>
<Tab name="kubectl">

1. List all available endpoints:

    ```bash
    kubectl get virtualservice --all-namespaces
    ```

2. To trigger the recreation of their configuration, delete the Pods of Istio Ingress Gateway:

     ```bash
     kubectl delete pod -l app=istio-ingressgateway -n istio-system
     ```
     
</Tab>

<Tab name="Kyma Dashboard">

1. Go to the `istio-system` namespace.
2. In the **Workloads** section, select **Pods**.
3. Use the search function to filter for all Pods labeled with `app=istio-ingressgateway`.
4. To trigger the recreation of their configuration, delete each of the displayed Pods.
   ![Delete Pods with `app=istio-ingressgateway` label](../../assets/delete-istio-ingressgateway-pods.svg)

</Tab>
</Tabs>

If the restart doesn't help, follow these steps:

1. Check all ports used by Istio Ingress Gateway:

   <Tabs>
   <Tab name="kubectl">

   1. List all the Pods of Istio Ingress Gateway:

      ```bash
      kubectl get pod -l app=istio-ingressgateway -n istio-system -o name
      ```

   2. Replace `{ISTIO_INGRESS_GATEWAY_POD_NAME}` with a name of a listed Pod and check the ports that the `istio-proxy` container uses:

      ```bash
      kubectl get -n istio-system pod {ISTIO_INGRESS_GATEWAY_POD_NAME} -o jsonpath='{.spec.containers[*].ports[*].containerPort}'
      ```
   
   </Tab>
   <Tab name="Kyma Dashboard">

   1. Go to the `istio-system` namespace.
   2. In the **Workloads** section, select **Pods**.
   3. Search for a Pod labeled with `app=istio-ingressgateway` and click on its name.
   ![Search for a Pod with `app=istio-ingressgateway` label](../../assets/search-for-istio-ingress-gateway.svg)
   4. Scroll down to find the `Containers` section and check which ports the `istio-proxy` container uses.
   ![Check ports used by istio-proxy](../../assets/check-istio-proxy-ports.svg)

   </Tab>
   </Tabs>


2. If the ports `80` and `443` are not used, check the logs of the `istio-proxy` container for errors related to certificates.

   <Tabs>
   <Tab name="kubectl">
   
   Run the following command:

   ```bash
   kubectl logs -n istio-system -l app=istio-ingressgateway -c istio-proxy
   ```
   
   </Tab>
   <Tab name="Kyma Dashboard">

   Click **View Logs**.
   ![View logs of the istio-proxy-container](../../assets/view-istio-proxy-logs.svg)
   
   </Tab>
   </Tabs>

3. To make sure that a corrupted certificate is regenerated, verify if the **spec.enableKymaGateway** field of your APIGateway custom resource is set to `true`. If you are running Kyma provisioned through Gardener, follow the [Gardener troubleshooting guide](https://kyma-project.io/docs/kyma/latest/04-operation-guides/troubleshooting/security/sec-01-certificates-gardener/) instead.