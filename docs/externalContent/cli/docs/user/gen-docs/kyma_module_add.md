# kyma module add

Add a module.

## Synopsis

Use this command to add a module.

```bash
kyma module add <module> [flags]
```

## Flags

```text
  -c, --channel string          Name of the Kyma channel to use for the module
      --cr-path string          Path to the custom resource file
      --default-cr              Deploys the module with the default CR
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
