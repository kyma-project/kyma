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

## Periodic job for volume snapshots

You can create a CronJob to take periodic snapshots of PersistentVolumes. See a sample CronJob definition which includes the required ServiceAccount and roles:

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