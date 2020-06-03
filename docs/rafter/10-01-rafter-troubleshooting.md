---
title: AssetGroup processing fails due to duplicated default buckets
type: Troubleshooting
---

It may happen that the processing of a (Cluster)AssetGroup CR fails due to too many buckets with the `rafter.kyma-project.io/access:public` label.

To fix this issue, manually remove all default buckets with the `rafter.kyma-project.io/access:public` label:  

1. Remove the cluster-wide default bucket:

   ```bash
   kubectl delete clusterbuckets.rafter.kyma-project.io --selector='rafter.kyma-project.io/access=public'
   ```
2. Remove buckets from the Namespaces where you use them:

   ```bash
   kubectl delete buckets.rafter.kyma-project.io --selector='rafter.kyma-project.io/access=public' --namespace=default
   ```
This allows the (Cluster)AssetGroup controller to recreate the default buckets successfully.
