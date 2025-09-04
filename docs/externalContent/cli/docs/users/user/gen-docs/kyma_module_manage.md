# kyma module manage

Sets the module to the managed state.

## Synopsis

Use this command to set an existing module to the managed state.

```bash
kyma module manage <module> [flags]
```

## Flags

```text
      --policy string           Sets a custom resource policy (Possible values: CreateAndDelete, Ignore) (default "CreateAndDelete")
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma module](kyma_module.md) - Manages Kyma modules
