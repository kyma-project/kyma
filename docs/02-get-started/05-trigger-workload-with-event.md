---
title: Trigger a workload with an event
---

We already know how to create and expose a workload ([Function](03-deploy-expose-function.md) and [microservice](04-deploy-expose-microservice.md)). 
Now it's time to actually use them.
We're going to trigger a workload with an event. 

## Create a Function 

To trigger our workload, we need something to trigger it with. 
For this, we are going to create a new Function with a corresponding OAuth2Client and an APIRule. As this is a sample Function, we'll make it so that it both sends and receives the events. 

<div tabs name="Deploy a Function" group="trigger-workload">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:
<!--TODO: 1) Replace `{LASTORDER_FUNCTION}` below with the actual Function's code. -->

```bash
cat <<EOF | kubectl apply -f -
  {LASTORDER_FUNCTION}
---
  apiVersion: hydra.ory.sh/v1alpha1
  kind: OAuth2Client
  metadata:
    name: lastorder
  spec:
    grantTypes:
      - "client_credentials"
    scope: "read write"
    secretName: lastorder-oauth
---
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: lastorder
  spec:
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - path: /function
        methods: ["GET", "POST"]
        accessStrategies:
          - handler: oauth2_introspection
            config:
              required_scope: ["read"]
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
oauth2client.hydra.ory.sh/lastorder created
apirule.gateway.kyma-project.io/lastorder created
```

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. From the left navigation, go to **Functions** and click to create a new Function.
2. Name the Function `lastorder` and click **Create**.
3. In the inline editor for the Function, modify its source replacing it with this code: 
   <!--TODO: Replace `{LASTORDER_FUNCTION}` below with the actual Function's code. -->
    ```js
    {LASTORDER_FUNCTION}
    ```
4. In your Function's view, go to the **Configuration** tab.
5. Click on **Expose Function +**.
6. Provide the **Name** (`lastorder`) and **Subdomain** (`lastorder`) and click **Create**.
    > **NOTE:** Alternatively, from the left navigation go to **APIRules**, click on **Create apirules +**, and continue with step 3, selecting the appropriate **Service** from the dropdown menu.
5. Using the left navigation, switch to **Configuration** > **OAuth2 Clients**. 
6. Click to add a new OAuth2 Client. 
7. Provide the following parameters:
    - **Name**: `lastorder`
    - **Scopes**: `read`, `write`
    - **Grant types**: check `Client Credentials`

   _Optionally_, provide a custom Secret name: `lastorder-oauth`.
8. Click **Create**.
   
  </details>
</div>

## Create a Subscription

Next, to subscribe to an event so that we can actually listen for it, we need a [Subscription](../05-technical-reference/06-custom-resources/evnt-01-subscription.md) custom resource. We're going to be listening for an event of type `order.received.v1`. 

<div tabs name="Create a Subscription" group="trigger-workload">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run: 
<!--TODO: Make sure that `inapp` in the address in `value` below matches the value used in the `lastorder` Function, or replace it with whatever the Function's using. -->
```bash
cat <<EOF | kubectl apply -f -
  apiVersion: eventing.kyma-project.io/v1alpha1
  kind: Subscription
  metadata:
    name: orders-sub
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
          value: sap.kyma.custom.inapp.order.received.v1
    protocol: ""
    protocolsettings: {}
    sink: http://lastorder.default.svc.cluster.local
EOF
```

To check that the Subscription was created, run:
```bash
kubectl get subscriptions orders-sub -o=jsonpath="{.status.ready}"
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
    <!--TODO: Make sure that `inapp` in **Application name** and **Event version** below matches the value used in the `lastorder` Function, or replace it with whatever the Function's using. -->
    - **Application name**: `inapp`
    - **Event name**: `order.received`
    - **Event version**: `v1`
    - **Event version**: `inapp.order.received.v1`

    The name of the event Subscription is generated automatically and follows the `{FUNCTION_NAME}-{RANDOM_SUFFIX}` pattern.

  </details>
</div>

## Trigger the workload with an event

We created the `lastorder` Function, exposed it, and created a Subscription for it to listen for `order.received.v1` events. Now it's time to send such an event and trigger the Function.

In your Terminal, run: 
<!--TODO: Make sure that `inapp` in the **"ce-type"** address below matches the value used in the `lastorder` Function, or replace it with whatever the Function's using. -->
```bash
curl -v -X POST \
     -H "ce-specversion: 1.0" \
     -H "ce-type: sap.kyma.custom.inapp.order.received.v1" \
     -H "ce-source: Kyma" \
     -H "ce-eventtypeversion: v1" \
     -H "ce-id: 759815c3-b142-48f2-bf18-c6502dc0998f" \
     -H "content-type: application/json" \
     -d "{\"orderCode\":\"3211213\"}" \
     http://eventing-event-publisher-proxy.kyma-system/publish
```
<!--TODO: Check whether it works with the `http://eventing-event-publisher-proxy.kyma-system/publish` URL below and if not, what to put there (what the 'lastorder` Function's using) and replace it. -->

## Verify the event delivery

To verify that the event was properly delivered, call the Function: 

```bash
curl -ik "https://lastorder.$CLUSTER_DOMAIN"
```

On successful delivery, the call returns the order ID of the event and the information that it was shipped: 
<!-- TODO: Replace the response below with an actual response. It's something similar to the response below but I couldn't get it running so I don't have a response, I only got 204 No Content: 
```
> POST /publish HTTP/1.1
> Host: localhost:8081
> User-Agent: curl/7.64.1
> Accept: */*
> ce-specversion: 1.0
> ce-type: sap.kyma.custom.inapp.order.received.v1
> ce-source: /default/sap.kyma/tunas-prow
> ce-eventtypeversion: v1
> ce-id: 759815c3-b142-48f2-bf18-c6502dc0998f
> content-type: application/json
> Content-Length: 23
>
* upload completely sent off: 23 out of 23 bytes
< HTTP/1.1 204 No Content
< Date: Fri, 10 Sep 2021 14:14:30 GMT
<
```
-->
```bash
HTTP/2 200
access-control-allow-origin: *
content-length: 652
content-type: application/json; charset=utf-8
date: Mon, 13 Jul 2020 13:05:33 GMT
etag: W/"28c-MLZh1MyovyUrCPwMzfRWfVQwhlU"
server: istio-envoy
x-envoy-upstream-service-time: 991
x-powered-by: Express

[{"orderCode":"987654321","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
```
<!--TODO: Check that the description in this step matches the actual response "returns the order ID of the event and the information that it was shipped:" and if not, correct it. -->