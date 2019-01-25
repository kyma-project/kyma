# Ark

## Overview

Ark is a tool to back up and restore Kubernetes resources and persistent volumes. It can create backups on demand or on schedule, filter objects which should be backed up, and set TTL (time to live) for stored backups. For more details, see the official [Ark documentation](https://heptio.github.io/ark/v0.9.0/).

## Details

By default, Ark comes with GCP as a backup storage provider and no bucket set. With that configuration, the Ark server deployment scales down to 0 replicas because Ark cannot start without the proper configuration for the backup storage bucket. You can change this by providing proper credentials in the heptio-ark/ark secret and changing the configuration in config/default.

## Usage

### Prerequisites

- [Download](https://github.com/heptio/ark/releases) the ark binary and add it to the **$PATH** environment variable. 
- Create a lambda function in the `production` Namespace.

### Steps

1. List all the Namespaces in the Kyma cluster.

```
kubectl --kubeconfig="path_to_kubeconfig" get namespaces
``` 

2. List all the objects present in the `production` Namespace.

```
kubectl --kubeconfig="path_to_kubeconfig" get all,serviceinstance,servicebinding,servicebindingusage,function,subscription,api,eventactivation -n production
```

3. Run the `ark` command to become familiar with its usage and flags.

4. Create a backup of the `production` Namespace.

```
ark --kubeconfig="path_to_kubeconfig" backup create production-backup --include-namespaces production
```

5. List all the backups.

```
ark --kubeconfig="path_to_kubeconfig" get backups
```

6. View the description of the `production-backup`.

```
ark --kubeconfig="path_to_kubeconfig" backup describe production-backup
```

7. Optionally, you can check **ark logs** for any warnings or errors resulting from running the command from the previous step. 

```
kubectl --kubeconfig="path_to_kubeconfig" -n heptio-ark logs deploy/ark
``` 

8. Restore the backup and check the result.

```
ark --kubeconfig="path_to_kubeconfig" create restore --from-backup production-backup
```

```
ark --kubeconfig="path_to_kubeconfig" get restores
```

9. View all the restored resources.

```
kubectl --kubeconfig="path_to_kubeconfig" get all,serviceinstance,servicebinding,servicebindingusage,function,subscription,api,eventactivation -n production
```

> **WARNING:** Currently, the service catalog instances are not correctly restored, causing a limitation. Implementing a plugin may provide a solution to this issue. 