---
title: Manage your Applications using Kyma and Compass Console UI
type: Tutorial
---

This tutorial presents the basic flow in which you manually register an external Application's API into Compass and expose it into the Kyma Runtime. In the Kyma Runtime, you then create a lambda that calls the `order-service` API. While going through this tutorial, you will navigate between two UI views:
- Compass UI, where you create connections between Applications, Runtimes, and Scenarios
- Kyma Console UI, where you manage resources used in your Application, such as services, lambdas, and bindings

## Prerequisites

For simplicity reason, use the available Order Service as the sample external Application for this tutorial. Prepare the following:
- [`order-service`](./assets/order-service.yaml) file that contains the service definition, deployment, and its API
- [API specification](./assets/order-service-api-spec.yaml) of the `order-service`
- [Lambda function](./assets/lambda.yaml) that calls `order-service` for orders
- Kyma cluster with the Compass module enabled

>**NOTE:** Read [this](#installation-enable-compass-in-kyma-default-kyma-installation) document to learn how to install Kyma with the Compass module.

## Steps

### Deploy the external Application

1. Log in to the Kyma Console and create a new Namespace by selecting the **Add new namespace** button.
2. In the **Overview** tab in your Namespace, select the **Deploy new resource** button and use the [`order-service`](./assets/order-service.yaml) file to connect the Application.
3. In the **APIs** tab, copy the URL to your Application. You will use it in the Compass UI in the next steps.

### Register your Application in the Compass UI

1. Select **Back to Namespaces** in the top-left corner of the page and go to the **Compass** tab in the left navigation panel. It will navigate you to the Compass UI. Select a tenant you want to work on from the drop-down list on the top navigation panel. For the purpose of this tutorial, select the `default` tenant. In the **Runtimes** tab, there is already the default `kymaruntime` that you can work on to complete this tutorial.
2. Navigate to the **Application** tab in the main Console view and click **Create Application** to register your Application in Compass. For the purpose of this tutorial, name your Application `test-app`. By default, your Application and Runtime are assigned to the `DEFAULT` scenario.
3. Select `test-app` in the **Applications** view and add the API spec of the Order Service:
  a) Click the **+** button in **API Definitions** section and fill in all the required fields.
  b) Paste the URL to your Application in the **Target URL** field.
  c) Click the **Add specification** button and upload the `order-service` [API spec file](./assets/order-service-api-spec.yaml).
  d) In the **Credentials** tab, choose `None` from the drop-down list to select the credentials type. For the purpose of this tutorial, there is no need to secure the connection. Click **Create**.

### Use your Application in the Kyma Console UI

1. Go back to the Kyma Console UI. You can see that the `test-app` Application is registered in the **Applications** view. Select `test-app` and bind it to your Namespace by selecting the  **Create Binding** button.
2. Select your Namespace from the drop-down list and go to the **Catalog** view. See that your services are now available under the **Services** tab. Provision the service instance by choosing your service and clicking the **Add once** button in the top-right corner of the page.
3. Create a lambda function. In the **Overview** tab, click the **Deploy new resource** button and upload the `lambda.yaml` file.
4. Expose your lambda:
  a) Go to the **Lambdas** tab in the left navigation panel and choose the `call-order-service` lambda.
  b) In the **Settings & Code** section, click the **Select Function Trigger** button and expose your lambda via HTTPS.
  c) Untick the **Enable authentication** field as  there is no need to secure the connection for the purpose of this tutorial.
  d) Click **Add**.
  e) Scroll down to the end of your lambda view and bind your lambda to your instance by clicking the **Create Service Binding** button in the **Service Binding** section. Choose the ServiceInstance you want to bind your lambda to and click **Create Service Binding**.
  f) Save the settings in the right top right corner of the page.
  g) Click the **Lambdas** tab and wait until the lambda status is completed and marked as `1/1`.
5. Test your lambda. Navigate to your lambda and go to the **Testing** tab. After you click the **Send** button, you can see the following output in the **Response** field:
  ```
{
  "status": 200,
  "data": []
}
  ```

  You can test your lambda by performing the following actions in the **Payload** field:
    - `{"action":"add"}` - adds the new order
    - `{"action":"list"}` - lists all orders; this is the default command executed after you click the **Send** button
    - `{"action":"delete"}` - deletes all the orders

### Cleanup

Clean up your cluster after going through this tutorial. To do so, delete your resources in the following order:
1. Go to the **Lambdas** tab, unfold the vertical option menu and delete your lambda.
2. Go to the **Services** tab and delete `order-service`.
3. Go to the **Deployments** tab and delete the `order-service ` deployment.
4. Go to the **APIs** tab and delete the `order-service ` API.
5. Go to the **Instances** tab, navigate to **Services**, and deprovision your instance by selecting the trash bin icon.
6. Go to the **Overview** section and unbind `test-app` from your Namespace.
7. Go back to the Namespaces view and delete your Namespace.
8. In the Compass UI, remove `test-app` from the **Applications** view. If you go back to the **Applications** view in the Kyma Console UI, you can see that the `test-app` Application is removed. It can take a moment, refresh if the Application is still there.
