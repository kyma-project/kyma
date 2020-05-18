---
title: Back up Kyma
type: Installation
---
The user load on a Kyma cluster typically consists of various Kubernetes objects and volumes. Kyma relies on the managed Kubernetes cluster for periodic backups of the Kubernetes objects. That's why it does not require any manual setup to perform backups.

For example, Gardener uses `etcd` as the Kubernetes' backing store for all cluster data. This means all Kubernetes objects are stored on `etcd`. Gardener uses periodic jobs to take major and minor snapshots of the `etcd` database. A major snapshot including all the resources takes place every day, and each minor snapshot including only the changes takes place every five minutes. In case the `etcd` database experiences any problems, Gardener automatically restores the Kubernetes cluster using the latest snapshot.

However, volumes are typically not a part of these backups. That's why it is recommended to take periodic backups of your volumes. You can do this using the VolumeSnapshot Kubernetes API resource. Read the following sections to learn how to use it.

## On-demand volume snapshots

Kubernetes provides an API resource called VolumeSnapshot that you can use to take the snapshot of a volume on Kubernetes. You can then use the snapshot either to provision a new volume (pre-populated with the snapshot data) or to restore the existing volume to a previous state (represented by the snapshot).

Volume snapshot support is only available for [CSI drivers](https://kubernetes-csi.github.io/docs/), however, not all CSI drivers support it. You can find a list of all the drivers [here](https://kubernetes-csi.github.io/docs/drivers.html).

Follow this [tutorial](/root/kyma#tutorials-create-on-demand-volume-snapshots) to create on-demand volume snapshots for different providers. 

## Periodic job for volume snapshots

Users can create a CronJob to take snapshots of PersistentVolumes periodically.

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