---
title: Manage your Applications using Kyma and Compass Console UI
type: Tutorial
---

This tutorial presents end-to-end use case scenario that shows how to connect an external Application to Compass so that its lambda calls and gets the order services. While going through this tutorial, you will navigate between two UI views:
- Compass UI, where you create connections between Applications, Runtimes, and Scenarios
- Kyma Console UI, where you manage resources used in your Application, such as services, lambdas, and bindings

## Prerequisites

Use the [HTTP DB Service](https://github.com/kyma-project/examples/tree/master/http-db-service) application to go through this tutorial. Prepare the following:
- HTTP DB Service [API definition](./assets/http-db-service.yaml)
- [`call-order-service`](./assets/lambda.yaml) lambda function
- Kyma cluster with the Compass module enabled

>**NOTE:** Read [this](#installation-enable-compass-in-kyma-compass-as-a-central-management-plane) document to learn how to install Kyma with the Compass module enabled.

## Steps

### Set up external Application



### Compass UI

1. Log in to the Kyma Console and select the **Compass** tab in the left navigation panel to navigate to the Compass UI. Select a tenant you want to work on from the drop-down list on the top navigation panel.
2. In the **Runtime** tab, click on the Runtime you want to work on and assign it to the `DEFAULT` scenario.
3. Navigate to the **Application** view and click **Create Application**. By default, your Application is assigned to the `DEFAULT` scenario.
4. Click on your Application name and **Add API** of the HTTP DB Service. In the **Credentials** tab, choose `None` as credential type. The **Target URL** is the URL to your Application. In case of HTTP DB Service, proceed with the next steps before completing this field.

### Kyma Console UI

5. Go back to the Kyma Console UI. You can see that your Application is registered in the **Applications** view.
6. Click on your Application name and **Create Binding** to a given Namespace.
7. In the **Overview** tab of your Namespace, click the **Deploy new resource** button. Add the deployment and lambda files.
8. Go to the **Services** tab, click on the `http-db-service` Application and expose its API. You'll get the Target URL that you need in the Compass Console UI. Copy the link and navigate to the Compass UI to finish the step of adding API to your Application.
9. Back in the Console UI, go to the **Catalog** view. Your services are now available under the **Services** tab.
10. Choose your service and create a ServiceInstance by clicking the **Add once** button.
11. Go to the **Lambdas** tab and choose the `call-order-service` lambda. Click the **Select Function Trigger** button and expose your lambda via HTTPS. Untick the **Enable authentication** field.
12. Scroll down in your lambda view and create a new ServiceBinding to bind your lambda to your instance. Remember to save the settings at top of the page.
13. Go to the **Testing** tab in your lambda view. Click the **Send** button. You can see a new order in the **Response** field.


### Cleanup

1. Go to the **Applications** section in the Kyma Console UI and navigate to your Application. Unbind the Application from your Namespace.
2. In the Compass UI, remove the `DEFAULT` scenario from your Runtime.
3. Go back to the Kyma Console UI and see that the Application is removed.
