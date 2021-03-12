---
title: Send an event through an application
type: Tutorials
---

This tutorial describes how to send an event in the Kyma Console UI. It shows how to integrate the [XF Addons](https://github.com/SAP-samples/xf-addons/tree/master/addons) with Kyma to use mocked external applications to send events to Kyma and trigger a Function.

## Deploy the XF Addons and provision the Marketing mock

The XF Addons are cluster-wide addons that give you access to 3 instances of mocks that simulate external applications sending events to Kyma.
Follow the steps to deploy XF Addons and add the marketing mock to your Namespace.

1. In the Kyma Console, go to **Cluster Addons**.
2. Click **Add new configuration**.
    * Provide the name for your configuration.
    * Provide the URL to the repository: `github.com/SAP-samples/xf-addons//addons/index.yaml` The `index.yaml` file is an addons manifest which includes Marketing, Cloud for Customer, and Commerce Cloud mocks.
    * Click **Add** to add the configuration.
3. Wait for the cluster addon to have the status `READY`.
4. Go to **Namespaces**.
5. Click **+ Add new Namespace** to create a new one.
6. Go to **{YOUR_NAMESPACE}** >  **Catalog** > **Addons**.
7. Select the mock you want to provision. For this example, use **SAP Marketing Cloud - Mock**.
8. Click **Add once** to deploy it in your Namespace.
9. Check if the mock is available under **Instances** > **Add-Ons**. You can access it at the URL `marketing-{NAMESPACE}.{CLUSTER_DOMAIN}.`
10. When the mock is provisioned, an API Rule is also created so that the mocked application is available externally. When you go to **{YOUR_NAMESPACE}** > **API Rules** you can see the direct link to the mocked application.

## Connect a mocked application to Kyma

Follow these steps to connect the marketing mock to Kyma.

1. Go to **Applications/Systems** and click **Add Application** to create an application in Kyma.
2. Click **Create Binding** to bind the application to the Namespace where you will later provision the APIs provided by the mocks. You can use the `default` Namespace or create a new one for that purpose.
3. In the application, click **Connect Application**.
4. Copy the token.
5. Access the marketing mock at the URL `marketing-{NAMESPACE}.{CLUSTER_DOMAIN}.` or use the link under **{YOUR_NAMESPACE}** > **API Rules**.
6. Click **Connect**.
7. Paste the token and wait for the application to connect.
8. Register all or some APIs. Make sure you register **Business Events** to be able to send events.

    >**NOTE:** Local APIs are the ones available with the mock application. Remote APIs represent the ones registered in Kyma.

9. Registered APIs are available under **{YOUR_NAMESPACE}** > **Service Catalog** > **Services**
10. Click **Add once** to add the **Business Events API** Service to the Namespace.

##  Create Function

Follow the steps to create a Function that you will trigger with your event.

1. Go to the Namespace in which you deployed the **Business Events** API.
2. Go **Functions**.
3. Click **Create Function**
4. Go to **Configuration** tab to select the Event Trigger. For example, use **bo.interaction.created**. This will be the event that triggers your Function.
5. In the **Source** tab, add the code for your Function. For example:

    ```bash
    module.exports = {
        main: function (event, context) {
            console.log(event.data);
            return "Hello World!";
        }
    }
    ```

## Send events

Follow the steps to send an event to Kyma and trigger a Function.

1. In the mock application, go to **Remote APIs**.
2. Go to **SAP Marketing Cloud - Business Events**.
3. Select the Event, in this case **bo.interaction.created**. You can also add the event payload.
4. Send the event.
5. Check the logs in the **Logs** section under **Functions** to see if the event arrived.
