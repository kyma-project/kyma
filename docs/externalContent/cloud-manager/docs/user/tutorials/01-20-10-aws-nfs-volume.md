# Using NFS in Amazon Web Services

This tutorial explains how to use Network File System (NFS) to create ReadWriteMany (RWX) volumes in Amazon Web Services (AWS). You can use the created RWX volume from multiple workloads.

## Prerequisites

You have the Cloud Manager module added.

## Steps <!-- {docsify-ignore} -->

1. Create a namespace and export its value as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```
  
2. Create an AwsNfsVolume resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsNfsVolume
   metadata:
     name: my-vol
   spec:
     capacity: 100G
   EOF
   ```
  
3. Wait for the AwsNfsVolume to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready awsnfsvolume/my-vol --timeout=300s
   ```

   Once the newly created AwsNfsVolume is provisioned, you should see the following message:

   ```console
   awsnfsvolume.cloud-resources.kyma-project.io/my-vol condition met
   ```

4. Observe the generated PersistentVolume:

   ```shell
   kubectl -n $NAMESPACE get persistentvolume my-vol
   ```

   You should see the following details:

   ```console
   NAME     CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM            STORAGECLASS
   my-vol   100G       RWX            Retain           Bound    test-mt/my-vol               
   ```

   The `RWX` access mode allows the volume to be readable and writable from multiple workloads. The `Bound` status which means the PersistentVolumeClaim claiming this PV is created.

5. Observe the generated PersistentVolumeClaim:

   ```shell
   kubectl -n $NAMESPACE get persistentvolumeclaim my-vol
   ```

   You should see the following message:

   ```console
   NAME     STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS 
   my-vol   Bound    my-vol   100G       RWX                         
   ```

   Similarly to PV, it should have the `RWX` access mode and `Bound` status.

6. Create two workloads that both write to the volume:

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: my-script
   data:
     my-script.sh: |
       #!/bin/bash
       for xx in {1..20}; do 
         echo "Hello from \$NAME: \$xx" | tee -a /mnt/data/test.log
         sleep 1
       done
       echo "File content:"
       cat /mnt/data/test.log
       sleep 864000
   ---
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: awsnfsvolume-demo
   spec:
     selector:
       matchLabels:
         app: awsnfsvolume-demo
     replicas: 2
     template:
       metadata:
         labels:
           app: awsnfsvolume-demo
       spec:
         containers:
           - name: my-container
             image: ubuntu  
             command: 
               - "/bin/bash"
             args:
               - "/script/my-script.sh"
             env:
               - name: NAME
                 valueFrom:
                   fieldRef:
                     fieldPath: metadata.name
             volumeMounts:
               - name: my-script-volume
                 mountPath: /script
               - name: data
                 mountPath: /mnt/data
         volumes:
           - name: my-script-volume
             configMap:
               name: my-script
               defaultMode: 0744 
           - name: data
             persistentVolumeClaim:
               claimName: my-vol 
   EOF
   ```

   This workload should print a sequence of 20 lines to stdout and a file on the NFS volume.
   Then it should print the content of the file.

7. Print the logs of one of the workloads, run:

   ```shell
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=awsnfsvolume-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```

   The command should print something like:

   ```console
   ...
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 19
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 20
   File content:
   Hello from awsnfsvolume-demo-869c89df4c-8z9zl: 20
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 1
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 1
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 2
   Hello from awsnfsvolume-demo-869c89df4c-dsw97: 2
   Hello from awsnfsvolume-demo-869c89df4c-l8hrb: 3
   ...
   ```

   Note that the content after `File content:` contains prints from both workloads. This demonstrates the ReadWriteMany capability of the volume.

## Next Steps

To clean up, follow these steps:

1. Remove the created workloads:

   ```shell
   kubectl delete -n $NAMESPACE deployment awsnfsvolume-demo
   ```

2. Remove the created ConfigMap:

   ```shell
   kubectl delete -n $NAMESPACE configmap my-script
   ```

3. Remove the created AwsNfsVolume:

   ```shell
   kubectl delete -n $NAMESPACE awsnfsvolume my-vol
   ```

4. Remove the created default IpRange:

   > [!NOTE]
   > If you have other cloud resources using the default IpRange,
   > skip this step, and do not delete the default IpRange.

   ```shell
   kubectl delete -n kyma-system iprange default
   ```

5. Remove the created namespace:

   ```shell
   kubectl delete namespace $NAMESPACE
   ```
