---
title: Restore Troubleshooting 
type: Troubleshooting
---

## Pod stuck in Init phase

In case the `service-catalog-addons-service-binding-usage-controller` Pod gets stuck in the `Init` phase, try deleting the Pod:

```bash
kubectl delete $(kubectl get pod -l app=service-catalog-addons-service-binding-usage-controller -n kyma-system -o name) -n kyma-system
```

## Different DNS and public IP address

The [restore](/components/backup/#tutorial-restore-a-kyma-cluster) tutorial assumes that the DNS and the public IP values stay the same as for the backed up cluster. If they change for the new cluster, check the relevant fields in the Secrets and ConfigMaps overrides in the `kyma-installer` Namespace and update them with new values. Then run the installer again to propagate them to all the components:

```bash
kubectl label installation/kyma-installation action=install
```
