# kyma module

Manages Kyma modules.

## Synopsis

Use this command to manage modules in the Kyma cluster.

```bash
kyma module <command> [flags]
```

## Available Commands

```text
  add      - Add a module
  catalog  - Lists modules catalog
  delete   - Deletes a module
  list     - Lists the installed modules
  manage   - Sets the module to the managed state
  unmanage - Sets a module to the unmanaged state
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma](kyma.md)                                 - A simple set of commands to manage a Kyma cluster
* [kyma module add](kyma_module_add.md)           - Add a module
* [kyma module catalog](kyma_module_catalog.md)   - Lists modules catalog
* [kyma module delete](kyma_module_delete.md)     - Deletes a module
* [kyma module list](kyma_module_list.md)         - Lists the installed modules
* [kyma module manage](kyma_module_manage.md)     - Sets the module to the managed state
* [kyma module unmanage](kyma_module_unmanage.md) - Sets a module to the unmanaged state
