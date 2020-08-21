---
title: Connect an external application
type: Getting Started
---

Let's start with integrating an external application to Kyma. In this set of guides, we will use a mock application called [Commerce mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/commerce-mock-0.1.0) that is to simulate a monolithic application. You will learn how you can connect it to Kyma, and expose its API and events. We will subscribe to one of its events (**order.deliverysent.v1**) in other tutorials and use it to trigger the logic of a sample service and function.  

## Deploy the XF Addons and provision the Commerce mock

Commerce mock is a part of the XF Addons that are cluster-wide addons giving access to 3 instances of mocks that simulate external applications sending events to Kyma.
Follow these steps to deploy XF Addons and add the Commerce mock to your Namespace:

1. In the Kyma Console, go to **Cluster Addons**.
2. Click **Add New Configuration**.
    * Provide the name for your configuration.
    * In the **Urls** field, provide the link to the repository with the Addon: `github.com/SAP-samples/xf-addons/addons/index.yaml` The `index.yaml` file is an addons manifest with APIs of SAP Marketing Cloud, SAP Cloud for Customer, and SAP Commerce Cloud applications.
    * Click **Add** to add the configuration.
3. Wait for the cluster addon to have the status `READY`.
4. Go to **Namespaces** in the left navigation panel of the Console UI.
5. Click **+ Add new namespace** to create a new one, insert `test` name and select **Create** to confirm your changes.
6. Go to **{YOUR_NAMESPACE}** > **Catalog** > **Add-Ons**.
7. Select the mock you want to provision. For this example, use **[Preview] SAP Marketing Cloud - Mock**.
8. Click **Add once** to deploy it in your Namespace.
9. Check if the mock is available under **Instances** > **Add-Ons**. You can access it at `https://marketing-{NAMESPACE}.{CLUSTER_DOMAIN}.`

10. When the mock is provisioned, an API Rule is automatically created so that the mock application can be accessible outside the cluster. When you go to **{YOUR_NAMESPACE}** > **API Rules** and select the mock, you will the direct link to the mock application under **Host**.

## Connect the mock application to Kyma

After provisioning the mock, connect it to Kyma:

1. Back in the general Console UI view, go to **Applications/Systems** and click **Create Application**.
2. Open the newly created application placeholder and click **Connect Application**.
3. Copy the token and select **OK** to business close the pop-up box.
4. Access the SAP Marketing Cloud Mock mock at `https://marketing-{NAMESPACE}.{CLUSTER_DOMAIN}.` or use the link under **{YOUR_NAMESPACE}** > **API Rules**.
5. Click **Connect**.
6. Paste the token and wait for the application to connect.
7. Select **Register All** or just register **SAP Marketing Cloud - Business Events** to be able to send events.

    >**NOTE:** Local APIs are the ones available with the mock application. Remote APIs represent the ones registered in Kyma.

8. Back in your application view in the Console UI, select **Create Binding** to bind the application to your Namespace where you will later provision the APIs provided by the mocks.You can use the `default` Namespace or create a new one for that purpose.

_does it matter which Namespace? why not the same?_

9. Registered APIs are available under **{YOUR_NAMESPACE}** > **Service Catalog** > **Services**
10. Select the **SAP Marketing Cloud - Business Events** Service and click **Add once** to add it to the Namespace.
