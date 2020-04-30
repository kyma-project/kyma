---
title: Kyma Backup strategy
type: Details
---

The user load on a Kyma cluster typically consists of various Kubernetes objects and volumes. Kyma relies on the managed Kubernetes cluster for periodic backups of the Kubernetes objects. That's why it does not require you to set anything manually to perform backups.

For example, Gardener uses etcd as the Kubernetes' backing store for all cluster data. This means all Kubernetes objects are stored on etcd. Gardener uses periodic jobs to take major and minor snapshots of the etcd database. A major snapshot including all the resources takes place every day, and each minor snapshot including only the changes in between takes place every five minutes. In case the etcd database experiences any problems, Gardener automatically restores the Kubernetes cluster using the latest snapshot.

However, volumes are typically not a part of these backups. That's why it is recommended to take periodic backups of your volumes. You can do this using the VolumeSnapshot Kubernetes API resource. Read the following sections to learn how to use it.

## On-Demand Volume Snapshots

Kubernetes provides an API resource called VolumeSnapshot that you can use to take the snapshot of a volume on Kubernetes. You can then use the snapshot either to provision a new volume (pre-populated with the snapshot data) or to restore the existing volume to a previous state (represented by the snapshot).

VolumeSnapshot support is only available for [CSI drivers](https://kubernetes-csi.github.io/docs/), however, not all CSI drivers support the volume snapshot functionality. You can find a list of all the drivers with the supported functionalities [here](https://kubernetes-csi.github.io/docs/drivers.html).

As an example, assume that you have a `pvc-to-backup` PVC which you have created using a CSI-enabled StorageClass. You can trigger a snapshot by creating a VolumeSnapshot object like the following:

> **NOTE:** You must use CSI-enabled StorageClass to create a PVC, otherwise it won't be backed up.

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: volume-snapshot
spec:
  volumeSnapshotClassName: csi-snapshot-class
  source:
    persistentVolumeClaimName: pvc-to-backup
```

Now assume your PVC is corrupt, and you want to re-create it using the snapshot. Create it by using the snapshot you created before as the data source for the new PVC:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-restored
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: csi-storage-class
  resources:
    requests:
      storage: 10Gi
  dataSource:
    name: volume-snapshot
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
```

This will create a new `pvc-restored` PVC with pre-populated data from the snapshot.

For details about VolumeSnapshots, see [this](https://kubernetes.io/docs/concepts/storage/volume-snapshots/) document.

Follow the instructions below to enable this feature for various providers.

### AKS

Minimum Kubernetes version supported is 1.17.

1. Install CSI driver following [this documentation](https://github.com/kubernetes-sigs/azuredisk-csi-driver/blob/master/docs/install-csi-driver-master.md).

2. Follow [this example](https://github.com/kubernetes-sigs/azuredisk-csi-driver/tree/master/deploy/example/snapshot) to see how you can create a VolumeSnapshot.

### GKE

Minimum Kubernetes version supported is 1.14.

1. Enable the required feature gate on the cluster following [this](https://cloud.google.com/kubernetes-engine/docs/how-to/gce-pd-csi-driver) document.

2. Check out [this repository](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver) for the details on how to use VolumeSnapshots on GKE.

### Gardener

#### GCP

Gardener GCP uses CSI drivers by default as of Kubernetes 1.18, and it supports Volume snapshotting out of the box.

Create a VolumeSnapshotClass:

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

Create a VolumeSnapshot:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: snapshot
spec:
  source:
    persistentVolumeClaimName: <PVC_NAME>
```

Wait until `READYTOUSE` field gets `true` to verify that snapshotting gets succeeded:

```bash
kubectl get volumesnapshot -w
```

#### AWS

Gardener AWS uses CSI drivers by default as of Kubernetes 1.18, and it supports Volume snapshotting out of the box.

Create a VolumeSnapshotClass:

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

Create a VolumeSnapshot:

```yaml
apiVersion: snapshot.storage.k8s.io/v1beta1
kind: VolumeSnapshot
metadata:
  name: snapshot
spec:
  source:
    persistentVolumeClaimName: <PVC_NAME>
```

Wait until `READYTOUSE` field gets `true` to verify that snapshotting gets succeeded:

```bash
kubectl get volumesnapshot -w
```

#### Azure

Gardener Azure does not currently support CSI drivers, that's why you cannot use VolumeSnapshots. This support is planned for Kubernetes 1.19. For details, see [this issue](https://github.com/gardener/gardener-extension-provider-azure/issues/3).

### Periodic Job for Volume Snapshots

Users can create a Cronjob to take snapshots of the PersistentVolumes periodically.

Have a look at a sample CronJob with the required Service Account and roles:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: volume-snapshotter
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: volume-snapshotter
  namespace: <NAMESPACE>
rules:
- apiGroups: ["snapshot.storage.k8s.io"]
  resources: ["volumesnapshots"]
  verbs: ["create", "get", "list", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: volume-snapshotter
  namespace: <NAMESPACE>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: volume-snapshotter
subjects:
- kind: ServiceAccount
  name: volume-snapshotter
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: volume-snapshotter
  namespace: <NAMESPACE>
spec:
  schedule: "@hourly" #Run once an hour, beginning of hour
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: volume-snapshotter
          containers:
          - name: job
            image: eu.gcr.io/kyma-project/test-infra/alpine-kubectl:v20200310-5f52f407
            command:
              - /bin/bash
              - -c
              - |
                # Create volume snapshot with random name.
                RANDOM_ID=$(openssl rand -hex 4)

                cat <<EOF | kubectl apply -f -
                apiVersion: snapshot.storage.k8s.io/v1beta1
                kind: VolumeSnapshot
                metadata:
                  name: volume-snapshot-${RANDOM_ID}
                  namespace: <NAMESPACE>
                  labels:
                    "job": "volume-snapshotter"
                    "name": "volume-snapshot-${RANDOM_ID}"
                spec:
                  volumeSnapshotClassName: <SNAPSHOT_CLASS_NAME>
                  source:
                    persistentVolumeClaimName: <PVC_NAME>
                EOF

                # Wait until volume snapshot is ready to use.
                attempts=3
                retryTimeInSec="30"
                for ((i=1; i<=attempts; i++)); do
                    STATUS=$(kubectl get volumesnapshot volume-snapshot-${RANDOM_ID} -n <NAMESPACE> -o jsonpath='{.status.readyToUse}')
                    if [ "${STATUS}" == "true" ]; then
                        echo "Volume snapshot is ready to use."
                        break
                    fi

                    if [[ "${i}" -lt "${attempts}" ]]; then
                        echo "Volume snapshot is not yet ready to use, let's wait ${retryTimeInSec} seconds and retry. Attempts ${i} of ${attempts}."
                    else
                        echo "Volume snapshot is still not ready to use after ${attempts} attempts, giving up."
                        exit 1
                    fi
                    sleep ${retryTimeInSec}
                done

                # Delete old volume snapshots.
                kubectl delete volumesnapshot -n <NAMESPACE> -l job=volume-snapshotter,name!=volume-snapshot-${RANDOM_ID}
```

### Troubleshooting

#### VolumeSnapshot creation failed

If a PersistentVolumeClaim is not bound, the attempt to create a volume snapshot from that PersistentVolumeClaim will fail with no retries. An event will be logged to indicate that the PersistentVolumeClaim is not bound.

Note that this can happen if the PersistentVolumeClaim spec and the VolumeSnapshot spec are in the same YAML file. In this case, when the VolumeSnapshot object is created, the PersistentVolumeClaim object is created but volume creation is not complete and therefore the PersistentVolumeClaim is not yet bound. You must wait until the PersistentVolumeClaim is bound and then create the snapshot.
