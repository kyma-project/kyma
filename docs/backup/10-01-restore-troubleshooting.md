---
title: Restore troubleshooting
type: Troubleshooting
---

## Pod stuck in Init phase

In case the `service-catalog-addons-service-binding-usage-controller` Pod gets stuck in the `Init` phase, try deleting the Pod:

```bash
kubectl delete $(kubectl get pod -l app=service-catalog-addons-service-binding-usage-controller -n kyma-system -o name) -n kyma-system
```

## Different DNS and public IP address

The [restore](/components/backup/#tutorial-restore-a-kyma-cluster) tutorial assumes that the DNS and the public IP values for the new cluster are the same as for the backed up cluster. If they change, check the relevant fields in the Secrets and ConfigMaps overrides in the `kyma-installer` Namespace and update them with new values. Then run the installer again to propagate them to all the components:

```bash
kubectl -n default label installation/kyma-installation action=install
```

## Eventing does not work

Check if all NatssChannels are reporting as ready:

```bash
kubectl get natsschannels.messaging.knative.dev -n kyma-system
```

If one or more channels report as not ready, delete their corresponding channel services. These services will be automatically recreated by the controller.

```bash
kubectl delete service -l messaging.knative.dev/role=natss-channel -n kyma-system
kubectl annotate natsschannels.messaging.knative.dev -n kyma-system restore=done --all
```

## AssetGroup restoration fails due to duplicated default buckets

Velero restores Kyma resources asynchronously, following the alphabetical `kind` object order. This creates issues for Rafter as Velero first restores (Cluster)AssetGroups and then buckets. When the (Cluster)AssetGroup controller notices there are no default buckets available, it automatically tries to create them. If Velero restores the default buckets in the meantime, (Cluster)AssetGroups restoration will fail due to too many buckets with the `rafter.kyma-project.io/access:public` label.

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
