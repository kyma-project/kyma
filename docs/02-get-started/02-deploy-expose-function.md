---
title: Deploy and expose a Function
---

Now that you've installed Kyma, let's deploy your first Function. We'll call it `hello-world`.

## Create a function

First, let's create the Function and apply it.

<div tabs name="Deploy a Function" group="deploy-expose-function">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. In Kyma Dashboard, go to the `default` Namespace.
2. Go to **Workloads** > **Functions**.
3. Click on **Create Function**.
4. Name the Function `hello-world` and click **Create**.
  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
kyma init function --name hello-world
kyma apply function
```

  </details>
</div>


### Verify the Function deployment

Now let's make sure that the Function has been deployed successfully.

<div tabs name="Verify the Function deployment" group="deploy-expose-function">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

The operation was successful if the Function **Status** changed to `RUNNING`.

> **NOTE:** You might need to wait a few seconds for the status to change.
  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
kubectl get functions hello-world
```

The operation was successful if the statuses for **CONFIGURED**, **BUILT**, and **RUNNING** are `True`.


  </details>
</div>

## Expose the Function

After we've got our `hello-world` Function deployed, we might want to expose it outside our cluster so that it's available for other external services.

> **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always secure the workload you expose with [OAuth2](../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md) or [JWT](../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).

First, let's create an [APIRule](../05-technical-reference/00-custom-resources/apix-01-apirule.md) for the Function.

<div tabs name="Expose the Function" group="deploy-expose-function">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Go to **Discovery and Network** > **API Rules**.
2. Click on **Create API Rule**.
3. Provide the **Name** (`hello-world`).
4. From the **Service Name** dropdown, select `hello-world`.
5. Choose your host from the **Host** dropdown and replace the asterix (*) with the name of your subdomain (`hello-world`).

  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1beta1
  kind: APIRule
  metadata:
    name: hello-world
    namespace: default
  spec:
    gateway: kyma-system/kyma-gateway
    host: hello-world.$CLUSTER_DOMAIN
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
      name: hello-world
      port: 80
EOF
```

  </details>
</div>

### Verify the Function exposure

Now let's verify that the Function has been exposed successfully.

<div tabs name="Access the Function" group="deploy-expose-function">
  <details open>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

In your Function's **Configuration** tab, click on the APIRule's **Host**.
This opens the Function's external address as a new page.

> **NOTE:** Alternatively, from the left navigation go to **API Rules**, and click on the **Host** URL there.

The operation was successful if the page says `Hello World from the Kyma Function main running on nodejs14!`.
  </details>
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
curl https://hello-world.$CLUSTER_DOMAIN
```

The operation was successful if the call returns `Hello Serverless`.

  </details>
</div>
