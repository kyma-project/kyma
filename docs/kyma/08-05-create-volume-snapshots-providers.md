---
title: Create on-demand volume snapshots for cloud providers
type: Tutorials
---

These tutorials show how to create on-demand volume snapshots for cloud providers. Before you proceed with the tutorial, read the general instructions on [creating volume snapshots](/#tutorials-create-on-demand-volume-snapshots).

<div tabs name="backup-providers">
  <details>
  <summary label="AKS">
  Create a volume snapshot for AKS
  </summary>

## Prerequisites

The minimum supported Kubernetes version is 1.17.

## Steps

1. [Install the CSI driver](https://github.com/kubernetes-sigs/azuredisk-csi-driver/blob/master/docs/install-csi-driver-master.md).
2. [Create a volume snapshot](https://github.com/kubernetes-sigs/azuredisk-csi-driver/tree/master/deploy/example/snapshot).

  </details>
  <details>
  <summary label="GKE">
  Create a volume snapshot for GKE
  </summary>

## Prerequisites

The minimum supported Kubernetes version is 1.14.

## Steps

1. [Enable the required feature gate on the cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/gce-pd-csi-driver).
2. Check out [the repository for the Google Compute Engine Persistent Disk (GCE PD) CSI driver](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver) for details on how to use volume snapshots on GKE.

  </details>
</div>

## Create volume snapshots for Gardener providers

<div tabs name="backup">
  <details>
  <summary label="GCP">
  GCP
  </summary>

### Prerequisites

As of Kubernetes version 1.18, Gardener GCP uses CSI drivers by default and supports taking volume snapshots out of the box.

### Steps

1. Create a VolumeSnapshotClass:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshotClass
metadata:
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
  name: snapshot-class
driver: pd.csi.storage.gke.io
deletionPolicy: Delete
```

2. Create a VolumeSnapshot resource:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: snapshot
spec:
  source:
    persistentVolumeClaimName: {PVC_NAME}
```

3. Wait until the **READYTOUSE** field has the `true` status to verify that the snapshot was taken successfully:

```bash
kubectl get volumesnapshot -w
```

  </details>
  <details>
  <summary label="AWS">
  AWS
  </summary>

### Prerequisites

As of Kubernetes version 1.18, Gardener AWS uses CSI drivers by default and supports taking volume snapshots out of the box.

### Steps

1. Create a VolumeSnapshotClass:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshotClass
metadata:
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
  name: snapshot-class
driver: ebs.csi.aws.com
deletionPolicy: Delete
```

2. Create a VolumeSnapshot resource:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: snapshot
spec:
  source:
    persistentVolumeClaimName: {PVC_NAME}
```

3. Wait until the **READYTOUSE** field receives the `true` status to verify that the snapshot was taken successfully:

```bash
kubectl get volumesnapshot -w
```
  </details>
  <details>
  <summary label="Azure">
  Azure
  </summary>

Gardener Azure does not currently support CSI drivers, that's why you cannot use volume snapshots. This support is planned for Kubernetes 1.19. For details, see [this issue](https://github.com/gardener/gardener-extension-provider-azure/issues/3).

  </details>
</div>
