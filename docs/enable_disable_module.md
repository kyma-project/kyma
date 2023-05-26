---
title: Enable and disable a Kyma module
---

## Overview

Your cluster comes with the Kyma custom resource (CR) already installed. It collects all metadata about the cluster, such as enabled modules, their statuses, or synchronization, using Lifecycle Manager. Lifecycle Manager uses `moduletemplate.yaml` to enable or disable modules on your cluster. 

## Procedure

> **NOTE:** To quickly enable a module, go to your **Cluster Details** view and select **Add Module**. Then, choose the channel of your choice and click **Add** to choose and upload your module.

> **TIP:** You can also configure your Kyma CR to enable a module manually. For more details, see [Kyma CR](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/technical-reference/api/kyma-cr.md).

<div tabs name="steps" group="enable-module">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

Check which modules are available on your cluster. Run: 
   ```bash
   kyma alpha list module
   ```

You should get a result similar to this example:

   ```bash
   operator.kyma-project.io/module-name    Domain Name (FQDN)         Channel     Version                      Template                     State
        cluster-ip                   kyma-project.io/cluster-ip        fast       v0.0.24    kyma-system/moduletemplate-cluster-ip-fast   <no value>
   ```

Enable a module on your cluster in the release channel of your choice. Run: 

   ```bash
   kyma alpha enable module {MODULE_NAME} --channel {CHANNEL_NAME} --wait
   ```

You should see the following message:

```bash
   - Successfully connected to cluster
   - Modules patched!
   ```

Similarly, to disable a module, run: 

   ```bash
   kyma alpha disable module {MODULE_NAME}
   ``` 
You should see the same message as the one displayed when you enable a module.

</details>
<details>
<summary label= Kyma Dashboard>
Kyma Dashboard
</summary>

Follow these steps to enable a Kyma module in Kyma Dashboard:
1. Go to the `kyma-system` Namespace.
2. In the **Kyma** section, Choose the **Kyma** resource.
3. Select your Kyma instance (`default-kyma`) and click **Edit**.
4. In the **Modules** section, click **Add**.
5. Choose the name of your desired module.
6. _Optionally_, choose the available channel.
7. Select **Update**.

This process may take a while, depending on the number of modules. The operation was successful if the module Status changed to `READY`.

To disable a module, edit your Kyma instance and click on the trash icon next to your module, then force update the changes. Your module should disappear from the Module list.
</details>
</div>

To configure your module, use the module CR that you can find in the module repository. 