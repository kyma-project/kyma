# kyma alpha app

Manages applications on the Kubernetes cluster.

## Synopsis

Use this command to manage applications on the Kubernetes cluster.

```bash
kyma alpha app <command> [flags]
```

## Available Commands

```text
  push - Push the application to the Kubernetes cluster
```

## Flags

```text
  -h, --help                    Help for the command
      --kubeconfig string       Path to the Kyma kubeconfig file
      --show-extensions-error   Prints a possible error when fetching extensions fails
      --skip-extensions         Skip fetching extensions from the cluster
```

## See also

* [kyma alpha](kyma_alpha.md)                   - Groups command prototypes for which the API may still change
* [kyma alpha app push](kyma_alpha_app_push.md) - Push the application to the Kubernetes cluster
