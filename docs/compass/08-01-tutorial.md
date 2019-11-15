---
title: Use Compass
type: Tutorial
---

This tutorial present end-to-end use case scenario that shows how to manage your Application landscape in Compass.

## Prerequisites

Use [HTTP DB Service](https://github.com/kyma-project/examples/tree/master/http-db-service) to go through this tutorial. Download the following:
- HTTP DB Service [API definition](https://github.com/kyma-project/examples/blob/master/http-db-service/docs/api/api.yaml)
- HTTP DB Service [deployment file](https://github.com/kyma-project/examples/blob/master/http-db-service/deployment/deployment.yaml)
- [`call-order-service`](./assets/lambda.yaml) lambda function


## Steps

### Compass UI view

1. Install Kyma and enable Compass modules. Read [this](#installation-installation) document to learn how.   
2. Log in to the Kyma Console and navigate to the **Compass** UI tab. Select a tenant you want to work on.
3. In the **Runtime** tab, assign `DEFAULT` scenario to the existing Runtime.
4. Navigate to the **Application** view and **Create Application**. Assign your application to the `DEFAULT` scenario.
5. Click on your Application name and **Add API** of the HTTP DB Service. Choose `None` as credential type.
6. Target URL:
>In case of your own Application, just copy its url here.

### Kyma Console view
You can see that your Application is registered in the **Applications** view of the Kyma Console.
1. Click on your Application name and **Create Binding** to a given Namespace.
2. In the **Overview** tab, click the **Deploy new resource** button. Add the deployment and lambda files.
3. Go to the **Services** tab, click on you Application name and expose its API. You'll get the Target URL.
4. Go to the **Catalog** view. Your ServiceClasses is now available.
5. Create a ServiceInstace.
6. Go to the **Lambdas** tab. Click the **Select Function Trigger** button and expose your lambda via HTTPS. Untick the **Enable authentication** field.
7. Create a new Service Binding and bind your lambda to your instance. Remember to save the settings.
8. In the Instances view, get GATEWAY_URL credentials.

### Cleanup
* Remove the mapping to the Namespace from the Runtime UI.
* On the Compass UI, remove the `DEFAULT` scenario from the Runtime.
* The Application is removed from the Runtime UI.
