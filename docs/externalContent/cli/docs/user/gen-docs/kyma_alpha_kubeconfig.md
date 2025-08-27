# kyma alpha kubeconfig

Manages access to the Kyma cluster.

## Synopsis

Use this command to manage access to the Kyma cluster

```bash
kyma alpha kubeconfig <command> [flags]
```

## Available Commands

```text
  generate - Generate kubeconfig with a Service Account-based or oidc tokens
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md)                                         - Groups command prototypes for which the API may still change
* [kyma alpha kubeconfig generate](kyma_alpha_kubeconfig_generate.md) - Generate kubeconfig with a Service Account-based or oidc tokens
