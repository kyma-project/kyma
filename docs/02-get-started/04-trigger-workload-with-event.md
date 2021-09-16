---
title: Trigger a workload with an event
---

We already know how to create and expose a workload ([Function](02-deploy-expose-function.md) and [microservice](03-deploy-expose-microservice.md)). 
Now it's time to actually use them.
We're going to trigger a workload with an event. 

## Create a Function 

For this purpose, we are going to create a new Function and expose it with an APIRule. As this is a sample Function, we'll make it so that it both sends and receives the events. 

<div tabs name="Deploy a Function" group="trigger-workload">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

```bash
cat <<EOF | kubectl apply -f -
  apiVersion: serverless.kyma-project.io/v1alpha1
  kind: Function
  metadata:
    labels:
      serverless.kyma-project.io/build-resources-preset: local-dev
      serverless.kyma-project.io/function-resources-preset: S
      serverless.kyma-project.io/replicas-preset: S
    name: lastorder
    namespace: default
  spec:
    deps: '{ "dependencies": {}}'
    maxReplicas: 1
    minReplicas: 1
    source: |
      let lastEvent = {}
      module.exports = {
        main: async function (event, context) {
          if (event["ce-type"]) {
            console.log("Received event:", event.data)
            lastEvent = event.data
          }
          else {
            console.log("Returning last event: ", lastEvent)
          }
          
          return lastEvent 
        } 
      }
---
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: lastorder
    namespace: default
  spec:
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - path: /.*
        methods: ["GET", "POST"]
        accessStrategies:
          - handler: allow
            config: {}
    service:
      host: lastorder.$CLUSTER_DOMAIN
      name: lastorder
      port: 80
EOF
```

If the resources were created successfully, the command returns this message:

```bash
function.serverless.kyma-project.io/lastorder created
apirule.gateway.kyma-project.io/lastorder created
```

To check that the Function is properly exposed, call it: 

```bash
curl https://lastorder.$CLUSTER_DOMAIN
```

As we haven't sent our event yet, we expect to get an empty object `{}` in response.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. From the left navigation, go to **Functions** and click to create a new Function.
2. Name the Function `lastorder` and click **Create**.
3. In the inline editor for the Function, modify its source replacing it with this code:
    ```js
    let lastEvent = {}
    module.exports = {
      main: async function (event, context) {
        if (event["ce-type"]) {
          console.log("Received event:", event.data)
          lastEvent = event.data
        }
        else {
          console.log("Returning last event: ", lastEvent)
        }
        
        return lastEvent 
      } 
    }
    ```
4. In your Function's view, go to the **Configuration** tab.
5. Click on **Expose Function +**.
6. Provide the **Name** (`lastorder`) and **Subdomain** (`lastorder`) and click **Create**.
    > **NOTE:** Alternatively, from the left navigation go to **APIRules**, click on **Create apirules +**, and continue with step 3, selecting the appropriate **Service** from the dropdown menu.

To check that the Function is properly exposed, call it. 
In your Function's **Configuration** tab, click on the APIRule's **Hostname**. This opens the Function's external address as a new page. As we haven't sent our event yet, we expect the page to return an empty object `{}`.

  </details>
</div>

## Create a Subscription

Next, to subscribe to an event so that we can actually listen for it, we need a [Subscription](../05-technical-reference/00-custom-resources/evnt-01-subscription.md) custom resource. We're going to be listening for an event of type `order.received.v1`. 

<div tabs name="Create a Subscription" group="trigger-workload">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run: 
```bash
cat <<EOF | kubectl apply -f -
   apiVersion: eventing.kyma-project.io/v1alpha1
   kind: Subscription
   metadata:
     name: lastorder-sub
     namespace: default
   spec:
     filter:
       filters:
       - eventSource:
           property: source
           type: exact
           value: ""
         eventType:
           property: type
           type: exact
           value: sap.kyma.custom.myapp.order.received.v1
     sink: http://lastorder.default.svc.cluster.local
EOF
```

To check that the Subscription was created and is ready, run:
```bash
kubectl get subscriptions lastorder-sub -o=jsonpath="{.status.ready}"
```

The operation was successful if the returned status says `true`.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Using the left navigation, go back to **Workloads** > **Functions**.
2. Select your `lastorder` Function and navigate to the **Configuration** tab.
3. Click on **Add Event Subscription+**.
4. Provide the following parameters:
    - **Application name**: `myapp`
    - **Event name**: `order.received`
    - **Event version**: `v1`
    - **Event version**: `myapp.order.received.v1`

    The name of the event Subscription is generated automatically and follows the `{FUNCTION_NAME}-{RANDOM_SUFFIX}` pattern.

  </details>
</div>

## Trigger the workload with an event

We created the `lastorder` Function, exposed it, and created a Subscription for it to listen for `order.received.v1` events. Now it's time to send such an event and trigger the Function. For the sake of this example, we'll port-forward the Kyma Eventing Service to localhost. 

1. Port-forward the Kyma Eventing Service to localhost. We will use port `3000`. Run: 
   ```bash
   kubectl -n kyma-system port-forward service/eventing-event-publisher-proxy 3000:80
   ```
2. Now trigger your Function. In another Terminal window run: 

   ```bash
   curl -v -X POST \
        -H "ce-specversion: 1.0" \
        -H "ce-type: sap.kyma.custom.myapp.order.received.v1" \
        -H "ce-source: /default/io.kyma-project/custom" \
        -H "ce-eventtypeversion: v1" \
        -H "ce-id: 759815c3-b142-48f2-bf18-c6502dc0998f" \
        -H "content-type: application/json" \
        -d "{\"orderCode\":\"3211213\"}" \
        http://localhost:3000/publish
   ```

## Verify the event delivery

To verify that the event was properly delivered, call the Function: 

```bash
curl -ik "https://lastorder.$CLUSTER_DOMAIN"
```

On successful delivery, the call returns the order code of the last registered order: 

```bash
HTTP/2 200
x-powered-by: Express
access-control-allow-origin: *
content-type: application/json; charset=utf-8
content-length: 23
etag: W/"17-SZJZtmfwj0jkot+V3pinY0wAUWs"
date: Thu, 16 Sep 2021 13:05:03 GMT
x-envoy-upstream-service-time: 5
server: istio-envoy

{"orderCode":"3211214"}%
```