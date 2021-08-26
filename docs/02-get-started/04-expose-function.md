---
title: Expose the Function
---

Now that we've got our `hello-world` Function deployed, let's expose it outside our cluster.

## Create an APIRule

First, let's create an APIRule for the Function. 

<div tabs name="Expose the Function" group="expose-function">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: hello-world
    namespace: default
  spec:
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - accessStrategies:
        - config: {}
          handler: allow
        methods:
          - GET
          - POST
          - PUT
          - PATCH
          - DELETE
          - HEAD
        path: /.*
    service:
      host: hello-world
      name: hello-world
      port: 80
EOF
```

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In your Function's view, go to the **Configuration** tab.
2. Click on **Expose Function +**.
3. Provide the **Name** (`hello-world`) and **Hostname** (`hello-world`) and click **Create**.

> **NOTE:** Alternatively, from the left navigation go to **APIRules**, click on **Create apirules**, and continue with step 3.
  </details>
</div>

## Verify the Function exposure

Now let's verify that the Function has been exposed successfully.

<div tabs name="Access the Function" group="expose-function">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
curl https://hello-world.$CLUSTER_DOMAIN
```

The operation was successful if the call returns `Hello Serverless`.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

In your Function's **Configuration** tab, click on the APIRule's **Hostname**. 
This will open the Function's external address in a new page. 

The operation was successful if the page says `Hello World!`. 
  </details>
</div>
