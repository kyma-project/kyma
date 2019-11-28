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
kubectl label installation/kyma-installation action=install
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
