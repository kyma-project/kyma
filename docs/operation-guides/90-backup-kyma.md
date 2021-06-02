---
title: Back up Kyma
---
The Kyma cluster load consists of Kubernetes [objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) and [volumes](https://kubernetes.io/docs/concepts/storage/volumes/). 

Kyma relies on a managed Kubernetes cluster for periodic backups of Kubernetes _objects_ to avoid any manual steps.

>**CAUTION:** Automatic backup does not include Kubernetes _volumes_. Back up your volumes periodically either on demand, or set up a periodic job.

## Object backup

For example, Gardener uses etcd as the Kubernetes backing store for all cluster data. Gardener runs periodic jobs to take major and minor snapshots of the etcd database to include Kubernetes objects in the backup. The major snapshot that includes all resources is taken on a daily basis, and minor snapshots happen every five minutes. If the etcd database experiences any problems, Gardener automatically restores the Kubernetes cluster using the most recent snapshot.

## On-demand volume snapshots

You should back up your volumes periodically  with the [VolumeSnapshot API resource](https://kubernetes.io/docs/concepts/storage/volume-snapshots/#volumesnapshots), which is provided by Kubernetes.

Learn how to create volume snapshops 

You can use your snapshot to provision a new volume pre-populated with the snapshot data or to restore the existing volume to the state represented by the snapshot.

>**INFO:** Taking volume snapshots is possible thanks to [Container Storage Interface (CSI) drivers](https://kubernetes-csi.github.io/docs/) which allow third-party storage providers to expose storage systems in Kubernetes. For details on available drivers, see the [full list of drivers](https://kubernetes-csi.github.io/docs/drivers.html).



Follow the [tutorial](#tutorials-create-on-demand-volume-snapshots-for-cloud-providers) to create on-demand volume snapshots for cloud providers. 

>**TIP:** Follow the instructions on [restoring resources using Velero](#tutorials-restore-resources-using-velero) to learn how to back up and restore individual resources.
