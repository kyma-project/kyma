# Cannot Find Kyma Resource in Kyma Dashboard

## Symptom

When I access Kyma dashboard, I cannot find any Kyma resources in the left panel.

## Cause

The relevant Kyma module is not added. For example, if you're missing the Function resource, you must add the Serverless module first. In the case of Subscription, you must add the Eventing module. For the list of all available modules, see [Kyma Modules](https://help.sap.com/docs/btp/sap-business-technology-platform/kyma-modules?locale=en-US).

## Solution

1. Go to Kyma dashboard. The URL is in the Overview section of your subaccount.
2. Choose **Modify Modules -> Add**.
3. In the **Add Modules** section, check the modules you want to add, and select **Add**.
4. **Optional:** At the module level, you can overwrite the default release channel for the modules you are adding. Under the Advanced options, choose your preferred release channel.

This process may take a while, depending on the number of modules. The operation was successful when the module status changes to `READY`.
