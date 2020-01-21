---
title: Manage your Applications using Kyma and Compass Console UI
type: Tutorial
---

This tutorial presents the basic flow where user manually registers an external Application APIs into Compass and then exposes them into the Kyma Runtime. Next, in the Kyma Runtime user creates a lambda which calls the `order-service` API. While going through this tutorial, you will navigate between two UI views:
- Compass UI, where you create connections between Applications, Runtimes, and Scenarios
- Kyma Console UI, where you manage resources used in your Application, such as services, lambdas, and bindings

## Prerequisites

For simplicity reason, use the [HTTP DB Service](https://github.com/kyma-project/examples/tree/master/http-db-service) as the example external Application to go through this tutorial. Prepare the following:
- [`order-service`](./assets/order-service.yaml) file that contains the service definition, deployment, and Api
- `order-service` [API spec](./assets/order-service-api-spec.yaml)
- [Lambda function](./assets/lambda.yaml) that calls `order-service` for orders
- Kyma cluster with the Compass module enabled

>**NOTE:** Read [this](#installation-enable-compass-in-kyma-default-kyma-installation) document to learn how to install Kyma with the Compass module enabled.

## Steps

### Deploy the external Application

1. Log in to the Kyma Console and create a new Namespace by clicking the **Add new namespace** button.
2. In the **Overview** tab in your Namespace, click the **Deploy new resource** button and use the file with the  [`order-service`](./assets/order-service.yaml) to connect the Application.
3. In the **APIs** tab, copy the URL to your Application. You will use in the Compass UI in the next steps.

### Register your Application in the Compass UI

1. Go **Back to Namespaces** at the left top of the page and select the **Compass** tab in the left navigation panel, which navigates you to the Compass UI. Select a tenant you want to work on from the drop-down list on the top navigation panel. For the purpose of this tutorial, select the `default` tenant.
2. In the **Runtimes** tab, there is already the default `kymaruntime` that you can work on to complete this tutorial.
3. Navigate to the **Application** view and click **Create Application** to register your Application in Compass. For the purpose of this tutorial, name your Application `test-app`. By default, your Application and Runtime are assigned to the `DEFAULT` scenario.
4. Click on `test-app` in the **Applications** view and add the API spec of the HTTP DB Service. To do so, click the **+** button in **API Definitions** section and fill in all the required fields. Paste the URL to your Application in the **Target URL** field. Click the **Add specification** button below and upload the `order-service` [API spec file](./assets/order-service-api-spec.yaml). In the **Credentials** tab, choose `None` as credential type from the drop-down list since, for the purpose of this tutorial, there is no need to secure the connection. Click **Create**.

### Use your Application in the Kyma Console UI

1. Go back to the Kyma Console UI. You can see that the `test-app` Application is registered in the **Applications** view. Click on `test-app` and bind it to your Namespace by clicking the  **Create Binding** button. Select the previously created Namespace from the drop-down list.
2. Go to the **Catalog** view. See that your services are now available under the **Services** tab. Provision your instance by choosing your service and clicking the **Add once** button at the right top corner of the page.
3. Create a lambda function. To do so, go **Back to Namespaces** at the left top of the page and select your Namespace. In the **Overview** tab, click the **Deploy new resource** button and upload the `lambda.yaml` file.
4. Expose your lambda. Go to the **Lambdas** tab at the left navigation panel and choose the `call-order-service` lambda. In the **Settings & Code** section, click the **Select Function Trigger** button and expose your lambda via HTTPS. Untick the **Enable authentication** field since, for the purpose of this tutorial, there is no need to secure the connection. Click **Add**. Scroll down to the end of your lambda view and bind your lambda to your instance by clicking the **Create Service Binding** button in the **Service Binding** section. Choose the ServiceInstance you want to bind your lambda to and click **Create Service Binding**. Remember to save the settings at the right top of the page. Click on the **Lambdas** tab and wait until the lambda status is completed and marked as `1/1`.  
5. Test your lambda. Navigate to your lambda and go to the **Testing** tab. After you click the **Send** button, you can see the following output in the **Response** field:
```
{
  "status": 200,
  "data": []
}
```
You can test your lambda by performing the following actions in the **Payload** section:
  - `{"action":"add"}` - adds the new order
  - `{"action":"list}"` - lists all orders; this is the default command executed after you click the **Send** button
  - `{"action":"delete"}` - deletes all the orders

### Cleanup

Clean up your cluster after going through this tutorial. To do so, delete your resources in the following order:
1. Go to the **Lambdas** tab, unfold the vertical option menu and delete your lambda.
2. Go to the **Services** tab and delete `http-db-service`.
3. Go to the **Deployments** tab and delete the `http-db-service` deployment.
4. Go to the **APIs** tab and delete the `http-db-service` API.
5. Go to the **Instances** tab, navigate to **Services**, and deprovision your instance by clicking on the thrash bin icon.
6. Go to the **Overview** section and unbind the `test-app` from your Namespace.
7. Go back to the Namespaces view and delete your Namespace.
8. In the Compass UI, remove the `test-app` from the **Applications** view. If you go back to the **Applications** view in the Kyma Console UI, you can see that the `test-app` Application is removed.
