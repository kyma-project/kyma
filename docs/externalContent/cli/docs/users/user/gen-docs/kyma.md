# kyma

A simple set of commands to manage a Kyma cluster.

## Synopsis

Use this command to manage Kyma modules and resources on a cluster.

```bash
kyma <command> [flags]
```

## Available Commands

```text
  alpha      - Groups command prototypes for which the API may still change
  completion - Generate the autocompletion script for the specified shell
  help       - Help about any command
  module     - Manages Kyma modules
  version    - Displays the version of Kyma CLI
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md)           - Groups command prototypes for which the API may still change
* [kyma completion](kyma_completion.md) - Generate the autocompletion script for the specified shell
* [kyma module](kyma_module.md)         - Manages Kyma modules
* [kyma version](kyma_version.md)       - Displays the version of Kyma CLI
