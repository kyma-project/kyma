---
title: Manage your Applications using Kyma and Compass Console UI
type: Tutorial
---

This tutorial presents an end-to-end use case scenario that shows how to connect an external Application to Compass so that its lambda calls and gets the order services. While going through this tutorial, you will navigate between two UI views:
- Compass UI, where you create connections between Applications, Runtimes, and Scenarios
- Kyma Console UI, where you manage resources used in your Application, such as services, lambdas, and bindings

## Prerequisites

Use the [HTTP DB Service](https://github.com/kyma-project/examples/tree/master/http-db-service) Application to go through this tutorial. Prepare the following:
- [File](./assets/http-db-service-deployment.yaml) that contains the HTTP DB Service definition and deployment
- HTTP DB Service [API spec](./assets/http-db-api-spec.yaml)
- [`call-order-service`](./assets/lambda.yaml) lambda function
- Kyma cluster with the Compass module enabled

>**NOTE:** Read [this](#installation-enable-compass-in-kyma-compass-as-a-central-management-plane) document to learn how to install Kyma with the Compass module enabled.

## Steps

### Connect the external Application

1. Log in to the Kyma Console and create the `test` Namespace by clicking the **Add new namespace** button.
2. In the **Overview** tab in your Namespace, click the **Deploy new resource** button and select the file with HTTP DB Service and deployment to connect the Application.

### Register your Application in the Compass UI

1. Go **Back to Namespaces** at the left top of the page and select the **Compass** tab in the left navigation panel, which navigates you to the Compass UI. Select a tenant you want to work on from the drop-down list on the top navigation panel. For the purpose of this tutorial, select the `default` tenant.
2. In the **Runtimes** tab, there is already `kymaruntime` that you can work on to complete this tutorial. This is the default Runtime that is already assigned to the `DEFAULT` scenario. You can also create a new Runtime by clicking the **Create runtime** button. After creating the Runtime, navigate to it, click **Edit**, and assign it to the scenario you want to work on. To create scenarios, go to the **Scenarios** tab and click the **Create Scenario** button. While creating a new scenario, you can assign Runtimes and Applications to it from the drop-down lists.
3. Navigate to the **Application** view and click **Create Application** to register your Application in Compass. For the purpose of this tutorial, name your Application `test-app`. By default, your Application is assigned to the `DEFAULT` scenario.
4. Click on `test-app` in the **Applications** view and add the API spec of the HTTP DB Service. To do so, click the **+** button in **API Definitions** section. Fill in all the required fields and click the **Add specification** button below and upload the HTTP DB Service API spec file. Paste the URL to your Application in the **Target URL** field. In case of HTTP DB Service, proceed with the next steps before filling this field. In the **Credentials** tab, choose `None` as credential type from the drop-down list since, for the purpose of this tutorial, there is no need to secure the connection.

### Use your Application in the Kyma Console UI

1. Go back to the Kyma Console UI. You can see that the `test-app` Application is registered in the **Applications** view. Click on `test-app` and bind it to your Namespace by clicking the  **Create Binding** button. Select the previously created `test` Namespace from the drop-down list.
2. Expose your Application. To do so, go back to the **Namespaces** tab and select the `test` Namespace. Select the **Services** tab at the left low navigation panel and click on the `http-db-service` Application. In the **Exposed APIs** section, click the **Expose API** button. In the required **Host** field, type `test` as a hostname. Do not secure your API since, for the purpose of this tutorial, there is no need to secure the connection. Click **Save**. In the **Exposed APIs** section below, you'll get the URL that you need in the **Target URL** field in the Compass UI view. Copy the link and navigate to the Compass UI to finish the step of adding API spec to your Application.
3. Back in the Kyma Console UI, go to the **Catalog** view. See that your services are now available under the **Services** tab. Provision your instance by choosing your service and clicking the **Add once** button.
4. Create a lambda function. To do so, go **Back to Namespaces** at the left top of the page and select the `test` Namespace. In the **Overview** tab, click the **Deploy new resource** button and upload the `lambda.yaml` file.
5. Expose your lambda. Go to the **Lambdas** tab at the left navigation panel and choose the `call-order-service` lambda. In the **Settings & Code** section, click the **Select Function Trigger** button and expose your lambda via HTTPS. Untick the **Enable authentication** field since, for the purpose of this tutorial, there is no need to secure the connection. Click **Add**. Scroll down to the end of your lambda view and bind your lambda to your instance by clicking the **Create Service Binding** button in the **Service Binding** section. Choose the ServiceInstance you want to bind your lambda to and click **Create Service Binding**. Remember to save the settings at top of the page. Click on the **Lambdas** tab and wait until the lambda status is completed and marked as `1/1`.  
6. Test your lambda. Navigate to your lambda and go to the **Testing** tab. After you click the **Send** button, you can see the following output in the **Response** field, :
```
{
  "status": 200,
  "data": []
}
```
You can test your lambda by performing the following actions in the **Payload** section:
  - `"action":"add"` - adds the new order
  - `"action":"list"` - lists all orders; this is the default command executed after you click the **Send** button
  - `"action":"delete"` - deletes all the orders

### Cleanup

Clean up your cluster after going through this tutorial. To do so, delete your resources in the following order:
1. Go to the **Lambdas** tab, unfold the vartical option menu and delete your lambda.
2. Go to the **Services** tab and delete `http-db-service`.
3. Go to the **Deployments** tab and delete `http-db-service` deployment.
4. Go to the **APIs** tab and delete `http-db-service-test` API.
5. Go to the **Instances** tab and deprovision your instance.
6. Go to the **Overview** section and unbind the `test-app` from your Namespace.
7. Go back to the Namespaces view and delete the `test` Namespace.
8. In the Compass UI, remove the `test-app` from the **Applications** view. If you go back to the Kyma Console UI, you can see that the `test-app` Application is removed.
