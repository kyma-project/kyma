---
title: Send an event through an application
type: Tutorials
---

This tutorial describes how to send an event through a mock commerce application using the Kyma Console UI.

## Connect a mocked application to Kyma

Follow these steps to connect a mock application to Kyma.

1. Install Kyma locally or on a cluster. Read more in the [Kyma installation](https://kyma-project.io/docs/root/kyma/#installation-installation) section. Open Kyma on your browser.
2. Go to **Namespaces** and click on **+ Add new namespace**. Pick the namespace title and click on **Create**.
// TODO - how to provision the commerce mock???
3. Click **Create Binding** to bind the application to the Namespace where you will later provision the APIs provided by the mock. You can use the `default` Namespace or create a new one for that purpose.
4. In the application, click **Connect Application**.
5. Copy the token.
6. Access the marketing mock at `commerce-{NAMESPACE}.{CLUSTER_DOMAIN}.` or use the link under **{YOUR_NAMESPACE}** > **API Rules**.
7. Click **Connect**.
8. Paste the token and wait for the application to connect.
9. Register all or some APIs. Make sure you register **SAP Commerce Cloud - Eventss** to be able to send events.
10. Registered APIs are available under **{YOUR_NAMESPACE}** > **Service Catalog** > **Services**
11. Click **Add once** to add the **Business Events API** Service to the Namespace.

##  Create Function

Follow the steps to create a Function that you will trigger with your event.

1. Go to the Namespace in which you deployed the **Business Events** API.
2. Go **Functions**.
3. Click **Create Function**
4. Go to **Configuration** tab to select the Event Trigger. For example, use **order.created**. This will be the event that triggers your Function.
5. In the **Source** tab, add the code for your Function:

```bash
module.exports = { main: function (event, context) {
        console.log(`event = ${JSON.stringify(event.data)}`);
        console.log(`headers = ${JSON.stringify(event.extensions.request.headers)}`);
    } }
```

## Send events

Follow the steps to send an event to Kyma and trigger a Function.

1. In the mock application, go to **Remote APIs**.
2. Go to **SAP Commerce Cloud - Eventss**.
3. Select the Event, in this case **order.created**. You can also add the event payload.
4. Send the event.
5. Check the logs in the **Logs** section under **Functions** to see if the event arrived.

