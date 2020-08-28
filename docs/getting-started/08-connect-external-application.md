---
title: Connect an external application
type: Getting Started
---

Let's now integrate an external application to Kyma. In this set of guides, we will use a sample application called [Commerce mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/commerce-mock-0.1.0) that simulates a monolithic application. You will learn how you can connect it to Kyma, and expose its API and events. In further guides, you will subscribe to one of Commerce mock events (`order.deliverysent.v1`) and see how you can use it to trigger the logic of the `orders-service` microservice.  

## Reference

This guide demonstrates how [Application Connector](/components/application-connector/) works in Kyma. It allows you to securely connect external solutions to your Kyma cluster.

## Steps

### Deploy the XF addons and provision Commerce mock

Commerce mock is a part of the XF addons that are cluster-wide addons giving access to three instances of mock applications that simulate external applications sending events to Kyma.

Follow these steps to deploy XF addons and add Commerce mock to the `orders-service` Namespace:

<div tabs name="provision-mock" group="connect-external-application">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Provision an [AddonsConfiguration CR](/components/helm-broker/#custom-resource-addons-configuration) with the mock application set:

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

2. Check that the AddonsConfiguration CR was created. Its phase should state `Ready`:

  ```bash
  kubectl get addonsconfigurations xf-mocks -n orders-service -o=jsonpath="{.status.phase}"
  ```

3. Create the ServiceInstance CR with Commerce mock:

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

4. Check that the ServiceInstance CR was created. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```
   </details>
   <details>
   <summary label="console-ui">
   Console UI
   </summary>

1. Switch to the `orders-service` Namespace, and go to **Configuration** > **Addons** in the left navigation panel.
2. Select **Add New Configuration**.
3. Once the new box opens up, enter `github.com/sap/xf-addons/addons/index.yaml` in the **Urls** field. Change the addon name to `xf-mocks`.

   > **NOTE:** The `index.yaml` file is a manifest for APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.

4. **Add** the configuration.
5. Wait for the addon to have the `READY` status.
6. Go to **Service Management** > **Catalog** in the left navigation panel.
7. Switch to the **Add-Ons** tab and select **[Preview] SAP Commerce Cloud - Mock** as the application to provision.

 > **TIP:** You can also use the search box in the upper right corner of the Console UI to find Commerce mock.

8. Click **Add once** to deploy the application in the `orders-service` Namespace. Leave the `default` plan. Change the name to `commerce-mock`.
9. Select **Create** to confirm the changes.

You will be redirected to the **Catalog Management** > **Instances** > **commerce-mock** view. Wait for it to have the `RUNNING` status.

When Commerce mock is provisioned, an API Rule for it is automatically created. When you go to the **API Rules** view in the `orders-service` Namespace and select `commerce-mock`, you will see the direct link to it under **Host**.

</details>
</div>

### Create the Application and retrieve a token

After provisioning Commerce mock, connect it to Kyma to expose its APIs and events on the cluster.

First, you must create an Application CR and then retrieve the token required to connect Commerce mock to the created Application CR.

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

2. Check that the Application CR was created. Its phase should state `deployed`:

   ```bash
   kubectl get application commerce-mock -o=jsonpath="{.status.installationStatus.status}"
   ```
3. Get a token required to connect Commerce mock to the Application CR. To do that, create the [TokenRequest CR](/components/application-connector/#custom-resource-token-request). The CR name must match the name of the application for which you want to get the configuration details. Run this command:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: applicationconnector.kyma-project.io/v1alpha1
  kind: TokenRequest
  metadata:
    name: commerce-mock
  EOF
  ```

4. Fetch the TokenRequest CR you created to get the token from the status section:

   ```bash
   kubectl get tokenrequest commerce-mock -o=jsonpath="{.status.url}"
   ```
   >**NOTE:** If the response doesn't contain any content, wait for some time and run the command again.

   The system returns a response similar to this one:

   ```bash
   https://connector-service.{CLUSTER_DOMAIN}/v1/applications/signingRequests/info?token=h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
   ```

   Save this output with the token to the clipboard, as you will need it in the next steps.

   > **CAUTION:** The token included in the output is valid for 5 minutes.

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Return to the general view in the Console UI by selecting **Back to Namespaces**.
2. Go to **Integration** > **Applications/Systems** and select **Create Application**.
3. Set the application's name to `commerce-mock` and select **Create** to confirm the changes.

  Wait for the application to have the `SERVING` status.

4. Open the newly created application and select **Connect Application**.
5. Copy the whole URL string with the token by clicking **Copy to Clipboard** and select **OK** to close the pop-up box.

  > **CAUTION:** The token included in the URL is valid for 5 minutes.

</details>
</div>

### Connect events

To connect events from Commerce mock to the microservice, follow these steps:  

1. Once in the `order-service` Namespace, go to **Configuration** > **API Rules** > **commerce-mock** in the left navigation panel.
2. Open the link under the **Host** column to access Commerce mock.
3. Click **Connect**.
4. Paste the previously copied URL string with the token. Confirm by selecting **Connect** and wait until the application gets connected.
5. Select **Register All** on the **Local APIs** tab or just register **SAP Commerce Cloud - Events** to be able to send events.

Once registered, you will see all Commerce mock APIs and events available under the **Remote APIs** tab.

> **TIP:** Local APIs are the ones available within the mock application. Remote APIs represent the ones registered in Kyma.

### Expose events in the Namespace

To expose events in the `orders-service` Namespace, first create an ApplicationMapping CR to bind the application to the Namespace. Then, provision the Events API in the Namespace using the ServiceInstance CR.

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

2. List available ServiceClass CR in the `orders-service` Namespace. Under the `EXTERNAL-NAME` column, find one with the `sap-commerce-cloud-events-*` prefix.

   ```bash
   kubectl get serviceclasses -n orders-service
   ```
   Copy the full `EXTERNAL NAME` to an environment variable. See the example:

   ```bash
   export EVENTS_EXTERNAL_NAME="sap-commerce-cloud-events-58d21"
   ```

3. Provision the Events API in the `orders-service` Namespace by creating a ServiceInstance CR:

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

4. Check that the ServiceInstance CR was created. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock-events -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Back in the Console UI, go to **Integration** > **Applications/Systems** > **commerce-mock**.

2. Select **Create Binding** to bind the application to the Namespace in which you will later provision the APIs provided by Commerce mock.

3. Select the `orders-service` Namespace and click **Create**.

4. Go to the `orders-service` Namespace view and navigate to **Service Management** > **Catalog**. Once on the **Services** tab, find **SAP Commerce Cloud - Events** and select it.

   > **TIP:** You can also use the search in the upper right corner.

5. Select **Add once** to add the service to the Namespace.

6. When the box pops up, change **Name** to `commerce-mock-events`, and confirm the changes by selecting **Create**.

This way you provisioned the events in the Namespace.

You will be redirected to the **Catalog Management** > **Instances** > **commerce-mock-events** view. Wait for the Events API to have the `RUNNING` status.

</details>
</div>

After all these steps, our microservice running in the `orders-service` Namespace can finally consume events from Commerce mock.
