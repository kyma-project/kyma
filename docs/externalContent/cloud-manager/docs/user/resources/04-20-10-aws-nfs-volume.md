# AwsNfsVolume Custom Resource

The `awsnfsvolume.cloud-resources.kyma-project.io` custom resource (CR) describes the AWS EFS
instance that can be used as a ReadWriteMany (RWX) volume in the cluster. Once the 
Amazon Elastic File System (AWS EFS) instance is provisioned
in the underlying cloud provider subscription, also the corresponding PersistentVolume (PV) and
PersistentVolumeClaim (PVC) are created in RWX mode, so they can be used from multiple cluster workloads. 
To use it as a volume in the cluster workload, specify the workload volume of the `persistentVolumeClaim` type.
A created AwsNfsVolume can be deleted only where there are no workloads that 
are using it, and when PV and PVC are unbound. 

The AwsNfsVolume requires an IP address in each zone of the cluster. Those IP addresses are 
allocated from the [IpRange](./04-10-iprange.md). If the IpRange is not specified in the AwsNfsVolume
then the default IpRange is used. If a default IpRange does not exist, it is automatically created.
Manually create a non-default IpRange with specified CIDR and use it only in advanced cases of network topology 
when you want to be in control of the network segments to avoid range conflicts with other networks. 

Though AWS EFS is elastic in its capacity, you must specify the capacity field on the resource since 
it's a required field on the PV and PVC. The recommended value for capacity is the maximum capacity that you 
would need. 

You can specify the `PerformanceMode` and `Throughput` AWS EFS configuration options, but they are optional
and default to `generalPurpose` and `bursting`.

By default, the created PV and PVC have the same name as the AwsNfsVolume resource, but you can optionally
specify their names, labels and annotations if needed. If PV or PVC already exists with a name equal to the one
being created, the provisioned AWS EFS remains and the AwsNfsVolume is put into the `Error`state.

## Specification <!-- {docsify-ignore} -->

This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter                   | Type                | Description                                                                                                                                                                                                                         |
|-----------------------------|---------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **ipRange**                 | object              | Optional IpRange reference. If omitted, default IpRange will be used, if default IpRange does not exist, it will be created                                                                                                         |
| **ipRange.name**            | string              | Name of the existing IpRange to use.                                                                                                                                                                                                |
| **capacity**                | quantity            | Maximum capacity of the volume. For example: 1300, 800M, 900Mi, 10G, 100Gi, 1T, 10Ti... To learn more, read about [K8S quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/ ':target=_blank') |
| **performanceMode**         | string              | The EFS PerformanceMode configuration option. One of `generalPurpose`, `maxIO`. Defaults to `generalPurpose`.                                                                                                                       |
| **throughput**              | string              | The EFS Throughput configuration option. One of `bursting`, `elastic`. Defaults to `bursting`.                                                                                                                                      |
| **volume**                  | object              | The PersistentVolume options. Optional.                                                                                                                                                                                             |
| **volume.name**             | string              | The PersistentVolume name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                                                                             |
| **volume.labels**           | map\[string\]string | The PersistentVolume labels. Optional. Defaults to nil.                                                                                                                                                                             |
| **volume.annotations**      | map\[string\]string | The PersistentVolume annotations. Optional. Defaults to nil.                                                                                                                                                                        |
| **volumeClaim**             | object              | The PersistentVolumeClaim options. Optional.                                                                                                                                                                                        |
| **volumeClaim.name**        | string              | The PersistentVolumeClaim name. Optional. Defaults to the name of the AwsNfsVolume resource.                                                                                                                                        |
| **volumeClaim.labels**      | map\[string\]string | The PersistentVolumeClaim labels. Optional. Defaults to nil.                                                                                                                                                                        |
| **volumeClaim.annotations** | map\[string\]string | The PersistentVolumeClaim annotations. Optional. Defaults to nil.                                                                                                                                                                   |

**Status:**

| Parameter                         | Type       | Description                                                                                                                        |
|-----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------|
| **state** (required)              | string     | Signifies the current state of **CustomObject**. Its value can be either `Ready`, `Processing`, `Error`, `Warning`, or `Deleting`. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                                                               |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                                                              |
| **conditions.message**            | string     | Provides more details about the condition status change.                                                                           |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                                                                |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.                                         |
| **conditions.type**               | string     | Provides a short description of the condition.                                                                                     |

## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary AwsNfsVolume custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsNfsVolume
metadata:
  name: my-vol
spec:
  capacity: 10T
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
