# GcpNfsVolume Custom Resource

The `gcpnfsvolume.cloud-resources.kyma-project.io` custom resource (CR) describes the GCP Filestore
instance that can be used as a ReadWriteMany (RWX) volume in the cluster. Once the GCP Filestore instance is provisioned
in the underlying cloud provider subscription, also the corresponding PersistentVolume (PV) and
PersistentVolumeClaim (PVC) are created in the RWX mode, so they can be used from multiple cluster workloads.
To use the GcpNfsVolume CR as a volume in the cluster workload, specify the workload volume of the `persistentVolumeClaim` type.
A created GcpNfsVolume can be deleted only where there are no workloads that
are using it, and when PV and PVC are unbound.

The zonal GCP Filestore requires 8 and the regional GCP Filestore requires 64 IP addresses. To learn
more, read [Configure a reserved IP address range](https://cloud.google.com/filestore/docs/creating-instances#configure_a_reserved_ip_address_range). 
Those IP addresses are
allocated from the [IpRange CR](./04-10-iprange.md). If an IpRange CR is not specified in the GcpNfsVolume,
then the default IpRange is used. If a default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified Classless Inter-Domain Routing (CIDR) and use it only in advanced cases of network topology
when you want to be in control of the network segments in order to avoid range conflicts with other networks.

You must specify the GcpNfsVolume capacity. Scalability and minimum/maximum capacity depend on the chosen GCP Filestore tier.
To learn more, read [GCP Filestore service tiers](https://cloud.google.com/filestore/docs/service-tiers).

You can optionally specify the `Tier` GCP Filestore configuration options. Its default value is `BASIC_HDD`, 
and can be one of `BASIC_HDD`, `BASIC_SSD`, `ZONAL` or `REGIONAL`. Legacy tiers are not supported.

By default, the created PV and PVC have the same name as the GcpNfsVolume resource, but you can optionally
specify their names, labels and annotations if needed. If PV or PVC already exist with the same name as the one
being created, the provisioned GCP Filestore remains and the GcpNfsVolume is put into the `Error` state.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                                                                                                                                                                                                                                                                                                 |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                 | object              | Optional IpRange reference. If omitted, default IpRange will be used, if default IpRange does not exist, it will be created.                                                                                                                                                                                                                                                                                |
| **ipRange.name**            | string              | Name of the existing IpRange to use.                                                                                                                                                                                                                                                                                                                                                                        |
| **capacityGb**              | int                 | Capacity of the provisioned volume in GiB. For a given tier range, a valid capacity must be between the minimum and maximum inclusive, and its difference from min must be divisible to scalability. For example, 1200 is not a valid capacity for the Zonal tier but 1280 (1024 + 256) is valid. To learn more, read [GCP Filestore service tiers](https://cloud.google.com/filestore/docs/service-tiers). |
| **location**                | string              | The zone (for zonal tiers) or the region (for regional tiers) option where GCP Filestore will reside. Defaults to the cluster region or one of its zones depending on the tier.                                                                                                                                                                                                                             |
| **tier**                    | string              | The GCP Filestore tier configuration option. Defaults to `BASIC_HDD`.                                                                                                                                                                                                                                                                                                                                       |
| **fileShareName**           | string              | The name of the file share option. Defaults to `vol1`. A file share is the directory on a Filestore instance where all shared files are stored.                                                                                                                                                                                                                                                             |
| **sourceBackup**            | object              | Optional source backup for restoring an existing backup while creating a new Filestore instance. The capacity of this filestore should be equal to or greater than the source Filestore of the backup. Tier limitations also apply. To learn more, read [GcpNfsVolumeRestore](./04-20-23-gcp-nfs-volume-restore.md).                                                                                               |
| **sourceBackup.name**       | object              | Name of the source GcpNfsVolumeBackup.                                                                                                                                                                                                                                                                                                                                                                      |
| **sourceBackup.namespace**  | object              | Namespace of the source GcpNfsVolumeBackup.                                                                                                                                                                                                                                                                                                                                                                 |
| **volume**                  | object              | The PersistentVolume options. Optional.                                                                                                                                                                                                                                                                                                                                                                     |
| **volume.name**             | string              | The PersistentVolume name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                                                                                                                                                                                                                                                     |
| **volume.labels**           | map\[string\]string | The PersistentVolume labels. Optional. Defaults to nil.                                                                                                                                                                                                                                                                                                                                                     |
| **volume.annotations**      | map\[string\]string | The PersistentVolume annotations. Optional. Defaults to nil.                                                                                                                                                                                                                                                                                                                                                |
| **volumeClaim**             | object              | The PersistentVolumeClaim options. Optional.                                                                                                                                                                                                                                                                                                                                                                |
| **volumeClaim.name**        | string              | The PersistentVolumeClaim name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                                                                                                                                                                                                                                                |
| **volumeClaim.labels**      | map\[string\]string | The PersistentVolumeClaim labels. Optional. Defaults to nil.                                                                                                                                                                                                                                                                                                                                                |
| **volumeClaim.annotations** | map\[string\]string | The PersistentVolumeClaim annotations. Optional. Defaults to nil.                                                                                                                                                                                                                                                                                                                                           |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **location**                      | string     | Signifies the location of the volume. This is particularly useful, if location is not provided in the spec.                        |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary GcpNfsVolume custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: GcpNfsVolume
metadata:
  name: my-vol
spec:
  capacityGb: 1024
  location: us-west1-a
---
apiVersion: v1
kind: Pod
metadata:
  name: workload
spec:
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: my-vol
  containers:
    - name: workload
      image: nginx
      volumeMounts:
        - mountPath: "/mnt/data1"
          name: data
```
