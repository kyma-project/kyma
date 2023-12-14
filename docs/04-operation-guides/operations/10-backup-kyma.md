# Back Up Kyma

## Context

The Kyma cluster load consists of Kubernetes [objects](https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/) and [volumes](https://kubernetes.io/docs/concepts/storage/volumes/).

### Object Backup

Kyma relies on a managed Kubernetes cluster for periodic backups of Kubernetes objects to avoid any manual steps.

>**CAUTION:** Automatic backup doesn't include Kubernetes volumes. Back up your volumes periodically either on demand, or set up a periodic job.

For example, Gardener uses [etcd](https://etcd.io/) as the Kubernetes backing store for all cluster data. Gardener runs periodic jobs to take major and minor snapshots of the etcd database to include Kubernetes objects in the backup.

The major snapshot that includes all resources is taken on a daily basis, while minor snapshots are taken every five minutes.

If the etcd database experiences any problems, Gardener automatically restores the Kubernetes cluster using the most recent snapshot.

### Volume Backup

We recommend that you back up your volumes periodically with the [VolumeSnapshot API resource](https://kubernetes.io/docs/concepts/storage/volume-snapshots/#volumesnapshots), which is provided by Kubernetes. You can use your snapshot to provision a new volume prepopulated with the snapshot data, or restore the existing volume to the state represented by the snapshot.

Taking volume snapshots is possible thanks to [Container Storage Interface (CSI) drivers](https://kubernetes-csi.github.io/docs/), which allow third-party storage providers to expose storage systems in Kubernetes. For details on available drivers, see the [full list of drivers](https://kubernetes-csi.github.io/docs/drivers.html).

You can create on-demand volume snapshots manually, or set up a periodic job that takes automatic snapshots periodically.

## Back Up Resources Using Velero

You can back up and restore individual resources manually or automatically with Velero. For more information, read the [Velero documentation](https://velero.io/docs/).
Be aware that a full backup of a Kyma cluster isn't supported. Start with the existing Kyma installation and restore specific resources individually.

## Create On-Demand Volume Snapshots

If you want to provision a new volume or restore the existing one, create on-demand [volume snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/):

<!-- tabs:start -->

#### **Gardener**

### Steps

  1. Create a VolumeSnapshotClass with the correct driver:
    - for GCP: `pd.csi.storage.gke.io`
    - for AWS: `ebs.csi.aws.com`
    - for Azure: `disk.csi.azure.com`

  ```yaml
  apiVersion: snapshot.storage.k8s.io/v1beta1
  kind: VolumeSnapshotClass
  metadata:
    annotations:
      snapshot.storage.kubernetes.io/is-default-class: "true"
    name: snapshot-class
  driver: <enter correct one for cloud provider>
  deletionPolicy: Delete
  ```
  
  2. Create a VolumeSnapshot resource:

  ```yaml
  apiVersion: snapshot.storage.k8s.io/v1beta1
  kind: VolumeSnapshot
  metadata:
    name: snapshot
  spec:
    volumeSnapshotClassName: snapshot-class
    source:
      persistentVolumeClaimName: {PVC_NAME}
  ```

  3. Wait until the **READYTOUSE** field has the `true` status to verify that the snapshot was taken successfully:

  ```bash
  kubectl get volumesnapshot -w
  ```

  4. Use this snapshot as a datasource to create a PVC:
  
  ```yaml
  apiVersion: v1
  kind: PersistentVolumeClaim
  metadata:
    name: pvc-restored
  spec:
    accessModes:
     - ReadWriteOnce
    resources:
      requests:
        storage: {SIZE_OF_ORIGINAL_PVC}
    dataSource:
      name: snapshot
      kind: VolumeSnapshot
      apiGroup: snapshot.storage.k8s.io
  ```

#### **AKS**

### Steps

  1. [Install the CSI driver](https://github.com/kubernetes-sigs/azuredisk-csi-driver/blob/master/docs/install-csi-driver-master.md).
  2. Follow our instructions to create a volume snapshot on Gardener, using the driver for Azure.

#### **GKE**

### Steps

  1. [Enable the required feature gate on the cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/gce-pd-csi-driver).
  2. Check out [the repository for the Google Compute Engine Persistent Disk (GCE PD) CSI driver](https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver) for details on how to use volume snapshots on GKE.

<!-- tabs:end -->

## Create a Periodic Snapshot Job

You can also create a CronJob to handle taking volume snapshots periodically. A sample CronJob definition that includes the required ServiceAccount and roles looks as follows:

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
  namespace: {NAMESPACE}
rules:
- apiGroups: ["snapshot.storage.k8s.io"]
  resources: ["volumesnapshots"]
  verbs: ["create", "get", "list", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: volume-snapshotter
  namespace: {NAMESPACE}
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
  namespace: {NAMESPACE}
spec:
  schedule: "@hourly" #Run once an hour, beginning of hour
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: volume-snapshotter
          containers:
          - name: job
            image: europe-docker.pkg.dev/kyma-project/prod/tpi/k8s-tools:v20231026-aa6060ec
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
                  namespace: {NAMESPACE}
                  labels:
                    "job": "volume-snapshotter"
                    "name": "volume-snapshot-${RANDOM_ID}"
                spec:
                  volumeSnapshotClassName: {SNAPSHOT_CLASS_NAME}
                  source:
                    persistentVolumeClaimName: {PVC_NAME}
                EOF

                # Wait until volume snapshot is ready to use.
                attempts=3
                retryTimeInSec="30"
                for ((i=1; i<=attempts; i++)); do
                    STATUS=$(kubectl get volumesnapshot volume-snapshot-${RANDOM_ID} -n {NAMESPACE} -o jsonpath='{.status.readyToUse}')
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
                kubectl delete volumesnapshot -n {NAMESPACE} -l job=volume-snapshotter,name!=volume-snapshot-${RANDOM_ID}
```
