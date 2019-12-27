---
title: Manage your Applications using Kyma and Compass Console UI
type: Tutorial
---

This tutorial present end-to-end use case scenario that shows how to manage an external application using Compass, so that its lambda calls and gets the order services.

## Prerequisites

Use the [HTTP DB Service](https://github.com/kyma-project/examples/tree/master/http-db-service) application to go through this tutorial. Download the following:
- HTTP DB Service [API definition](https://github.com/kyma-project/examples/blob/master/http-db-service/docs/api/api.yaml)
- HTTP DB Service [deployment file](https://github.com/kyma-project/examples/blob/master/http-db-service/deployment/deployment.yaml)
- [`call-order-service`](./assets/lambda.yaml) lambda function
- Kyma cluster with the Compass module enabled

>**NOTE:** Read [this](#installation-enable-compass-in-kyma-compass-as-a-central-management-plane) document to learn how to install Kyma with the Compass module enabled.

## Steps

### Compass UI view

1. Log in to the Kyma Console and navigate to the **Compass** UI which opens a new tab with the Compass UI Console view. Select a tenant you want to work on.
2. In the **Runtime** tab, click on the Runtime you want to work on and assign it to the `DEFAULT` scenario.
3. Navigate to the **Application** view and click **Create Application**. Assign your application to the `DEFAULT` scenario.
4. Click on your Application name and **Add API** of the HTTP DB Service. Choose `None` as credential type. The **Target URL** is the URL to your Application. In case of HTTP DB Service, proceed with the next steps before completing this field.


### Kyma Console view

Go back to the Kyma Console UI. You can see that your Application is registered in the **Applications** view.
1. Click on your Application name and **Create Binding** to a given Namespace.
2. In the **Overview** tab of your Namespace, click the **Deploy new resource** button. Add the deployment and lambda files.
3. Go to the **Services** tab, click on the `http-db-service` Application and expose its API. You'll get the Target URL that you need in the Compass Console UI.
4. Go to the **Catalog** view. Your services are now available.
5. Choose your service and create a ServiceInstance by clicking the **Add once** button.
6. Go to the **Lambdas** tab and choose the `call-order-service` lambda. Click the **Select Function Trigger** button and expose your lambda via HTTPS. Untick the **Enable authentication** field.
7. Create a new Service Binding and bind your lambda to your instance. Remember to save the settings in your lambda view.


### Cleanup

* Remove the mapping to the Namespace from the Runtime UI.
* On the Compass UI, remove the `DEFAULT` scenario from the Runtime.
* The Application is removed from the Runtime UI.
