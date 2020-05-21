---
title: AssetGroup restore fails due to duplicated default buckets
type: Troubleshooting
---

[Velero restores Kyma resources asynchronously](/root/kyma/#tutorials-restore-with-velero), following the alphabetical `kind` object order. This creates issues for Rafter as Velero first restores (Cluster)AssetGroups and then buckets. When the (Cluster)AssetGroup controller notices there are no default buckets available, it automatically tries to create them. If Velero restores the default buckets in the meantime, (Cluster)AssetGroups restoration will fail due to too many buckets with the `rafter.kyma-project.io/access:public` label.

To fix this issue, manually remove all default buckets with the `rafter.kyma-project.io/access:public` label right after restoring Kyma:  

1. Remove the cluster-wide default bucket:

   ```bash
   kubectl delete clusterbuckets.rafter.kyma-project.io --selector='rafter.kyma-project.io/access=public'
   ```

2. Remove buckets from the Namespaces where you use them:

   ```bash
   kubectl delete buckets.rafter.kyma-project.io --selector='rafter.kyma-project.io/access=public' --namespace=default
   ```

This allows the (Cluster)AssetGroup controller to recreate the default buckets successfully.
