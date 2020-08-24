---
title: Connect an external application
type: Getting Started
---

Let's now integrate an external application to Kyma. In this set of guides, we will use a mock application called [Commerce mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/commerce-mock-0.1.0) that is to simulate a monolithic application. You will learn how you can connect it to Kyma, and expose its API and events. We will subscribe to one of its events (**order.deliverysent.v1**) in other guides and see how you can use it to trigger the logic of our sample microservice.  

## Steps

### Deploy the XF addons and provision Commerce mock

Commerce mock is a part of the XF addons that are cluster-wide addons giving access to three instances of mocks that simulate external applications sending events to Kyma.

Follow these steps to deploy XF addons and add Commerce mock in the `orders-service` Namespace:

<div tabs name="provision-mock" group="connect-external-application">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Provision an AddonsConfiguration CR with the mocks:

```bash
cat <<EOF | kubectl apply -f  -
apiVersion: addons.kyma-project.io/v1alpha1
kind: AddonsConfiguration
metadata:
name: xf-mocks
namespace: orders-service
spec:
repositories:
- url: github.com/sap/xf-addons//addons/index.yaml
EOF
```
   > **NOTE:** The `index.yaml` file is an addons manifest with APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.

2. Check if the AddonsConfiguration CR was created. The CR phase should state `Ready`:

  ```bash
  kubectl get addonsconfigurations xf-mocks -n orders-service -o=jsonpath="{.status.phase}"
  ```

3. Create the ServiceInstance CR with the mock:

```bash
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

4. Check if the ServiceInstance CR was created. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```
   </details>
   <details>
   <summary label="console-ui">
   Console UI
   </summary>

1. Go to the `orders-service` Namespace in the Console UI and navigate to **Addons** under the **Configuration** section in the left navigation panel.
2. Select **Add New Configuration**.
3. Once the new box opens up, enter `github.com/sap/xf-addons//addons/index.yaml` in the **Urls** field. The addon name is automatically generated.

   > **NOTE:** The `index.yaml` file is an addons manifest with APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.

4. **Add** the configuration.
5. Wait for the addon to have the `READY` status.
6. Go to the **Catalog** view under the **Service Management** section in the left navigation panel.
7. Switch to the **Add-Ons** tab and select **[Preview] SAP Commerce Cloud - Mock** as the application to provision.

 > **TIP:** You can also use the search box in the upper right corner of the Console UI to find the mock.

8. Click **Add once** to deploy the application in the `orders-service` Namespace. Leave the `default` plan. The mock name will be automatically generated.
9. Select **Create** to confirm the changes.

You will be redirected to the **Catalog Management** > **Instances** > **{GENERATED_MOCK_NAME}** view. Wait for the mock to have the `RUNNING` status.

When Commerce mock is provisioned, an API Rule for it is automatically created. When you go to the **API Rules** view in the `orders-service` Namespace and select the mock, you will see the direct link to Commerce mock under **Host**.

</details>
</div>

### Connect Commerce mock to Kyma

After provisioning the mock, connect it to Kyma to expose its APIs and events on the cluster:

#### Create the Application and retrieve a token

First create the Application CR and then retrieve the token required to connect Commerce mock to the created Application. Follow these steps:

<div tabs name="create-application" group="connect-external-application">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Apply the Application CR definition to the cluster:

```bash
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

2. Check if the Application CR was created. The CR phase should state `deployed`:

   ```bash
   kubectl get application commerce-mock -o=jsonpath="{.status.installationStatus.status}"
   ```
3. Get a token required to connect Commerce mock to the Application CR. To do that, create the TokenRequest CR. The CR name must match the name of the application for which you want to get the configuration details. Run this command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: TokenRequest
metadata:
  name: commerce-mock
EOF
```

4. Fetch the TokenRequest CR you created to get the token from the status section. Run this command:

   ```bash
   kubectl get tokenrequest commerce-mock -o=jsonpath="{.status.url}"
   ```
   >**NOTE:** If the response doesn't contain any content, wait for some time and run the command again.

   A successful call should return a similar response:

   ```bash
   https://connector-service.{CLUSTER_DOMAIN}/v1/applications/signingRequests/info?token=h31IwJiLNjnbqIwTPnzLuNmFYsCZeUtVbUvYL2hVNh6kOqFlW9zkHnzxYFCpCExBZ_voGzUo6IVS_ExlZd4muQ==
   ```

   Save this token to the clipboard, as you will need it in the next steps.

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Return to the general view in the Console UI by selecting **Back to Namespaces**.
2. Go to **Applications/Systems** under the **Integration** section and select **Create Application**.
3. Set the Application's name as `commerce-mock` and select **Create** to confirm the changes.

Wait for the Application to have the `SERVING` status.

4. Open the newly created Application and select **Connect Application**.
5. Copy the token by clicking **Copy to Clipboard** and select **OK** to close the pop-up box.

</details>
</div>

### Connect events

To connect events from Commerce mock to the microservice, follow these steps:  

1. Access Commerce mock at `https://commerce-orders-service.{CLUSTER_DOMAIN}` or use the link under **API Rules** in the **Configuration** section in the `order-service` Namespace. You can also access Commerce mock through the direct link to the mock application under the **Host** column.
2. Click **Connect**.
3. Paste the token, confirm by selecting **Connect**, and wait until the application gets connected.
4. Select **Register All** on the **Local APIs** tab or just register **SAP Commerce Cloud - Events** to be able to send events.

Once registered, you will see all Commerce mock APIs and events available under the **Remote APIs** tab.

    >**NOTE:** Local APIs are the ones available within the mock application. Remote APIs represent the ones registered in Kyma.

#### Expose events in the Namespace

To expose events in a Namespace, first create an ApplicationMapping CR in the cluster to bind an application to the Namespace. Then, provision the Events API in the Namespace by ServiceInstance CR. Follow the instructions:

<div tabs name="expose-events-in-namespace" group="connect-external-application">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create an ApplicationMapping CR and apply it to the cluster:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: ApplicationMapping
metadata:
  name: commerce-mock
  namespace: orders-service
EOF
```

2. List available ServiceClass CRs in the `orders-service` Namespace and find under the `EXTERNAL-NAME` column the one with the `sap-commerce-cloud-events-*` prefix.

   ```bash
   kubectl get serviceclasses -n orders-service
   ```
   Copy the full `EXTERNAL NAME` to an environment variable. See the example:

```bash
export EVENTS_EXTERNAL_NAME="sap-commerce-cloud-events-58d21"
```

3. Provision the Events API in the `orders-service` Namespace by creating a ServiceInstance CR:

```bash
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

4. Check if the ServiceInstance CR was created. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance commerce-mock-events -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Back in the application view in the Console UI (**Integration** > **Applications/Systems** > **commerce-mock**), select **Create Binding** to bind the application to the Namespace in which you will later provision the APIs provided by the Commerce mock. Select `orders-service` Namespace and click **Create**.

2. Open the `orders-service` Namespace view and navigate to **Service Management** > **Catalog**. Once on the **Services** tab, find **SAP Commerce Cloud - Events** and select it.

   > **TIP:** You can also use the search in the upper right corner.

3. Select **Add once** to add the service to the Namespace.

4. When the box pops up, leave the default values and confirm the changes by selecting **Create**.

This way you provisioned the events (created ServiceClasses) in the Namespace.

You will be redirected to the **Catalog Manegement** > **Instances** > **{GENERATED_MOCK_NAME}** view. Wait for the Events API to have the `RUNNING` status.

</details>
</div>

After all these steps, our microservice running in the `orders-service` Namespace can consume events from Commerce mock.
