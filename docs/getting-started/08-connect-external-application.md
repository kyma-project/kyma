---
title: Connect an external application
type: Getting Started
---

Let's now connect an external application to Kyma. In this set of guides, we will use a sample application called [Commerce mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/commerce-mock-0.1.0) that simulates a monolithic application. You will learn how you can connect it to Kyma, and expose its API and events. In further guides, you will subscribe to one of Commerce mock events (`order.deliverysent.v1`) and see how you can use it to trigger the logic of the `orders-service` microservice.  

## Reference

This guide demonstrates how [Application Connector](/components/application-connector/) works in Kyma. It allows you to securely connect external solutions to your Kyma instance.

## Steps

### Deploy the XF addons and provision Commerce mock

Commerce mock is a part of the XF addons. These are cluster-wide addons giving access to three instances of mock applications, which in turn simulate external applications sending events to Kyma.

Follow these steps to deploy XF addons and add Commerce mock to the `orders-service` Namespace:

<div tabs name="provision-mock" group="connect-external-application">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Provision an [AddonsConfiguration custom resource (CR)](/components/helm-broker/#custom-resource-addons-configuration) with the mock application set:

  ```yaml
  cat <<EOF | kubectl apply -f  -
  apiVersion: addons.kyma-project.io/v1alpha1
  kind: AddonsConfiguration
  metadata:
    name: xf-mocks
    namespace: orders-service
  spec:
    repositories:
      - url: github.com/sap/xf-addons/addons/index.yaml
  EOF
  ```

   > **NOTE:** The `index.yaml` file is a manifest for APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.

2. Check that the AddonsConfiguration CR was created. This is indicated by the phase `Ready`:

  ```bash
  kubectl get addonsconfigurations xf-mocks -n orders-service -o=jsonpath="{.status.phase}"
  ```

3. Create a ServiceInstance CR with Commerce mock:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: servicecatalog.k8s.io/v1beta1
  kind: ServiceInstance
  metadata:
    name: commerce-mock
    namespace: orders-service
  spec:
    serviceClassExternalName: commerce-mock
    servicePlanExternalName: default
  EOF
  ```

4. Check that the ServiceInstance CR was created. This is indicated by the last condition in the CR status equal to `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

   </details>
   <details>
   <summary label="console-ui">
   Console UI
   </summary>

1. Switch to the `orders-service` Namespace. In the left navigation panel, go to **Configuration** > **Addons**.
2. Select **Add New Configuration**.
3. Once the new box opens up, enter `github.com/sap/xf-addons/addons/index.yaml` in the **Urls** field. Change the addon name to `xf-mocks`.

   > **NOTE:** The `index.yaml` file is a manifest for APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.

4. **Add** the configuration.
5. Wait for the addon to have the status `READY`.
6. In the left navigation panel, go to **Service Management** > **Catalog**.
7. Switch to the **Add-Ons** tab and select **[Preview] SAP Commerce Cloud - Mock** as the application to provision.

 > **TIP:** You can also use the search bar in the upper right corner of the Console UI to find Commerce mock.

8. Click **Add once** to deploy the application in the `orders-service` Namespace. Leave the `default` plan. Change the name to `commerce-mock`.
9. Select **Create** to confirm the changes.

You will be redirected to the **Service Management** > **Instances** > **commerce-mock** view. Wait for it to have the status `RUNNING`.

When Commerce mock is provisioned, a corresponding API Rule is automatically created. When you go to the **Discovery and Network** > **API Rules** view in the `orders-service` Namespace and select `commerce-mock`, you will see the direct link to it under **Host**.

  </details>
</div>

> **CAUTION:** If you have a Minikube cluster, you must add the [**spec.template.spec.hostAliases**](https://kubernetes.io/docs/concepts/services-networking/add-entries-to-pod-etc-hosts-with-host-aliases/) field in the **commerce-mock** Deployment with the following hostnames:
>
>  ```yaml
>  hostAliases:
>    - ip: $(minikube ip)
>      hostnames:
>        - connector-service.kyma.local
>        - gateway.kyma.local
>  ```

### Create the Application and retrieve a token

After provisioning Commerce mock, connect it to Kyma to expose its APIs and events to the cluster.

First, you must create an Application CR. Then, retrieve the token required to connect Commerce mock to the created Application CR.

Follow these steps:

<div tabs name="create-application" group="connect-external-application">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Apply the [Application CR](/components/application-connector/#custom-resource-application) definition to the cluster:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: Application
  metadata:
    name: commerce-mock
  spec:
    description: "Application for Commerce mock"
    labels:
      app: orders-service
      example: orders-service
  EOF
  ```

2. Check that the Application CR was created. This is indicated by the status `deployed`:

   ```bash
   kubectl get application commerce-mock -o=jsonpath="{.status.installationStatus.status}"
   ```

3. Get a token required to connect Commerce mock to the Application CR. To do that, create a [TokenRequest CR](/components/application-connector/#custom-resource-token-request). The CR name must match the name of the application for which you want to get the configuration details. Run this command:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: TokenRequest
  metadata:
    name: commerce-mock
  EOF
  ```

4. Fetch the TokenRequest CR you created to get the configuration URL with the token from the status section:

   ```bash
   kubectl get tokenrequest commerce-mock -o=jsonpath="{.status.url}"
   ```
   >**NOTE:** If the response doesn't contain any content, wait for a few moments and run the command again.

   The system returns a response similar to this one:

   ```bash
   https://connector-service.{CLUSTER_DOMAIN}/v1/applications/signingRequests/info?token=h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
   ```

   Save this URL with the token to the clipboard, as you will need it in the next steps.

   > **CAUTION:** The token included in the output is valid for 5 minutes.

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Return to the general view in the Console UI by selecting **Back to Namespaces**.
2. Go to **Integration** > **Applications/Systems** and select **Create Application**.
3. Set the application's name to `commerce-mock` and select **Create** to confirm the changes.

  Wait for the application to have the status `SERVING`.

4. Open the newly created application and select **Connect Application**.
5. Copy the whole URL string with the token by clicking **Copy to Clipboard** and select **OK** to close the pop-up box.

  > **CAUTION:** The token included in the URL is valid for 5 minutes.

  </details>
</div>

### Connect events

To connect events from Commerce mock to the microservice, follow these steps:  

1. Once in the `order-service` Namespace, go to **Discovery and Network** > **API Rules** > **commerce-mock** in the left navigation panel.
2. Open the link under the **Host** column to access Commerce mock.
3. Click **Connect**.
4. Paste the previously copied URL with the token, check **Insecure Connection**, confirm by selecting **Connect**, and wait until the application gets connected.
5. Select **Register All** on the **Local APIs** tab or just register **SAP Commerce Cloud - Events** to be able to send events.

Once registered, you will see all Commerce mock APIs and events available under the **Remote APIs** tab.

> **TIP:** Local APIs are the ones available within the mock application. Remote APIs are the ones registered in Kyma.

### Expose events in the Namespace

To expose events in the `orders-service` Namespace, first create an ApplicationMapping CR to bind the application to the Namespace. Then, enable the events API in the Namespace using the ServiceInstance CR.

Follow these steps:

<div tabs name="expose-events-in-namespace" group="connect-external-application">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create an [ApplicationMapping CR](/components/application-connector/#custom-resource-application-mapping) and apply it to the cluster:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: ApplicationMapping
  metadata:
    name: commerce-mock
    namespace: orders-service
  EOF
  ```

2. List the available ServiceClass CRs in the `orders-service` Namespace:

   ```bash
   kubectl get serviceclasses -n orders-service
   ```
   Under the `EXTERNAL-NAME` column, find one with the `sap-commerce-cloud-events-*` prefix. Copy the full `EXTERNAL NAME` to an environment variable. See the example:

   ```bash
   export EVENTS_EXTERNAL_NAME="sap-commerce-cloud-events-58d21"
   ```

3. Enable the events in the `orders-service` Namespace by creating a ServiceInstance CR:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: servicecatalog.k8s.io/v1beta1
  kind: ServiceInstance
  metadata:
    name: commerce-mock-events
    namespace: orders-service
  spec:
    serviceClassExternalName: $EVENTS_EXTERNAL_NAME
    servicePlanExternalName: default
  EOF
  ```

4. Check that the ServiceInstance CR was created. This is indicated by the last condition in the CR status equal to `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock-events -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Back in the Console UI, go to **Integration** > **Applications/Systems** > **commerce-mock**.

2. Select **Create Binding** to bind the application to the Namespace in which you will later enable the APIs provided by Commerce mock.

3. Select the `orders-service` Namespace and click **Create**.

4. Go to the `orders-service` Namespace view and navigate to **Service Management** > **Catalog**. Once on the **Services** tab, find **SAP Commerce Cloud - Events** and select it.

   > **TIP:** You can also use the search bar in the upper right corner.

5. Select **Add once** to add the service to the Namespace.

6. When a box pops up, change **Name** to `commerce-mock-events`. Confirm the changes by selecting **Create**.

This way you enabled the events in the Namespace.

You will be redirected to the **Service Management** > **Instances** > **commerce-mock-events** view. Wait for the events API to have the status `RUNNING`.

</details>
</div>

After all these steps, our microservice running in the `orders-service` Namespace can finally consume events from Commerce mock.
