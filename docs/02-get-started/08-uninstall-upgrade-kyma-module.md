# Uninstall and Upgrade Kyma with a Module

This guide shows how to quickly uninstall or upgrade Kyma with specific modules.

> [!NOTE]
> This guide describes uninstallation and upgrade of standalone Kyma with specific modules. If you are using SAP BTP, Kyma runtime (SKR), read [Enable and Disable a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?locale=en-US&version=Cloud) instead.

## Uninstall Kyma with a Module

You uninstall Kyma with a module using the `kubectl delete` command.

1. Find out the paths for the module you want to disable; for example, from the [Install Kyma with a module](./01-quick-install.md#steps) document.

2. Delete the module configuration:

   ```bash
   kubectl delete {PATH_TO_THE_MODULE_CUSTOM_RESOURCE}
   ```

3. To avoid leaving some resources behind, wait for the module custom resource deletion to be complete.

4. Delete the module manager:

   ```bash
   kubectl delete {PATH_TO_THE_MODULE_MANAGER_YAML_FILE}
   ```

## Upgrade a Kyma Module

To upgrade a Kyma module to the latest version, run the same `kubectl` commands used for its [installation](./01-quick-install.md).

## Related Links

To see the list of all available Kyma modules, go to [Kyma modules](../06-modules/README.md).
