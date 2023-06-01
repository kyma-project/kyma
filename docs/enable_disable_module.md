---
title: Enable and disable a Kyma module
---

## Overview

Your cluster comes with the Kyma custom resource (CR) already installed. It collects all metadata about the cluster, such as enabled modules, their statuses, or synchronization. The `moduletemplate.yaml` file contains the ModuleTemplate CR used to enable or disable modules on your cluster. 

## Enable a module

<div tabs name="Enable a module" group="enable-disable-module">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

1. Check which modules are available on your cluster. Run: 
   ```bash
   kyma alpha list module
   ```

   You should get a result similar to this example:

   ```bash
   operator.kyma-project.io/module-name    Domain Name (FQDN)         Channel     Version                      Template                     State
        cluster-ip                   kyma-project.io/cluster-ip        fast       v0.0.24    kyma-system/moduletemplate-cluster-ip-fast   <no value>
   ```

2. Enable a module on your cluster in the release channel of your choice. Run: 

   ```bash
   kyma alpha enable module {MODULE_NAME} --channel {CHANNEL_NAME} --wait
   ```

   You should see the following message:

   ```bash
   - Successfully connected to cluster
   - Modules patched!
   ```

</details>
<details>
<summary label= Kyma Dashboard>
Kyma Dashboard
</summary>

> **NOTE:** To quickly enable a module using Kyma Dashboard, go to your **Cluster Details** view and select **Add Module**. Then, select the channel of your choice and click **Add** to choose and upload your module.

Follow these steps to enable a Kyma module in Kyma Dashboard:
1. Go to the `kyma-system` Namespace.
2. In the **Kyma** section, Choose the **Kyma** resource.
3. Select your Kyma instance (`default-kyma`) and click **Edit**.
4. In the **Modules** section, click **Add**.
5. Choose the name of your desired module.
6. _Optionally_, choose the available channel.
7. Select **Update**.

This process may take a while, depending on the number of modules. The operation was successful if the module Status changed to `READY`.
</details>
</div>

> **TIP:** You can also configure the Kyma CR to enable a module manually. For more details, see [Kyma CR](https://github.com/kyma-project/lifecycle-manager/blob/main/docs/technical-reference/api/kyma-cr.md).

## Disable a module

<div tabs name="Disable a module" group="enable-disable-module">
  <details>
  <summary label="cli">
  Kyma CLI
  </summary>

To disable a module, run: 

   ```bash
   kyma alpha disable module {MODULE_NAME}
   ``` 
You should see the following message:

```bash
   - Successfully connected to cluster
   - Modules patched!
   ```

</details>
<details>
<summary label= Kyma Dashboard>
Kyma Dashboard
</summary>

Follow these steps to disable a Kyma module in Kyma Dashboard:
1. Go to the `kyma-system` Namespace.
2. In the **Kyma** section, Choose the **Kyma** resource.
3. Select your Kyma instance (`default-kyma`) and click **Edit**.
4. Click on the thrash icon next to your module and update the changes.

Your module should disappear from the Module list.
</details>
</div>

To configure your module, use the module CR that you can find in the module repository. 