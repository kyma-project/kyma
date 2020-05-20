---
title: Back up Kyma
type: Installation
---
The user load on a Kyma cluster consists of Kubernetes [objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) and [volumes](https://kubernetes.io/docs/concepts/storage/volumes/). Kyma relies on the managed Kubernetes cluster for periodic backups of Kubernetes objects, so you don't have to perform any manual steps.

For example, Gardener uses etcd as the Kubernetes backing store for all cluster data. Gardener runs periodic jobs to take major and minor snapshots of the etcd database to include Kubernetes objects in the backup. A major snapshot including all the resources happens evert day, and  minor snapshot, including only the changes, happens every five minutes. If the etcd database experiences any problems, Gardener automatically restores the Kubernetes cluster using the most recent snapshot.

>**NOTE:** Kubernetes volumes are typically not a part of backups. That's why you should take periodic backups of your volumes using the VolumeSnapshot API resource.

## On-demand volume snapshots

Kubernetes provides the [VolumeSnapshot API resource](https://kubernetes.io/docs/concepts/storage/volume-snapshots/#volumesnapshots) that you can use to create a snapshot of a Kubernetes volume. You can use the snapshot to provision a new volume pre-populated with the snapshot data or to restore the existing volume to the state represented by the snapshot.

Volume snapshots are supported only by [Container Storage Interface (CSI) drivers](https://kubernetes-csi.github.io/docs/), but not all of them are compliant. For more details on compliant drivers, see the [full list of drivers](https://kubernetes-csi.github.io/docs/drivers.html).

Follow this [tutorial](#tutorials-create-volume-snapshots-providers) to create on-demand volume snapshots for various providers. 
