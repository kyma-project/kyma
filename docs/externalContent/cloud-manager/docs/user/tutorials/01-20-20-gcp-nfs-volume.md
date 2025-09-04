# Using Network File System in Google Cloud

This tutorial explains how to use Network File System (NFS) to create ReadWriteMany (RWX) volumes in Google Cloud. You can use the created RWX volume from multiple workloads.

## Steps <!-- {docsify-ignore} -->

1. Create a namespace and export its value as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```
  
2. Create an GcpNfsVolume resource.

   ```shell
   cat <<EOF | kubectl -n $NAMESPACE apply -f -
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpNfsVolume
   metadata:
     name: my-vol
   spec:
     location: us-west1-a
     capacityGb: 1024
   EOF
   ```
  
3. Wait for the GcpNfsVolume to be in the `Ready` state.

   ```shell
   kubectl -n $NAMESPACE wait --for=condition=Ready gcpnfsvolume/my-vol --timeout=300s
   ```

   Once the newly created GcpNfsVolume is provisioned, you should see the following message:

   ```console
   gcpnfsvolume.cloud-resources.kyma-project.io/my-vol condition met
   ```

4. Observe the generated PersistentVolume:

   ```shell
   export PV_NAME=`kubectl get gcpnfsvolume my-vol -n $NAMESPACE -o jsonpath='{.status.id}'`
   kubectl get persistentvolume $PV_NAME
   ```

   You should see the following details:

   ```console
   NAME       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                    STORAGECLASS
   {PV_NAME}  1Ti        RWX            Retain           Bound    {NAMESPACE_NAME}/my-vol                             
   ```

   Note the `RWX` access mode which allows the volume to be readable and writable from multiple workloads, and the
   the `Bound` status which means the PersistentVolumeClaim claiming this PV is created.

5. Observe the generated PersistentVolumeClaim:

   ```shell
   kubectl -n $NAMESPACE get persistentvolumeclaim my-vol
   ```

   You should see the following message:

   ```console
   NAME     STATUS   VOLUME     CAPACITY   ACCESS MODES   STORAGECLASS 
   my-vol   Bound    {PV_NAME}  1Ti        RWX                                                
   ```

   Similarly to PV, note the `RWX` access mode and `Bound` status.

6. Create two workloads that both write to the volume. Run:

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
     name: gcpnfsvolume-demo
   spec:
     selector:
       matchLabels:
         app: gcpnfsvolume-demo
     replicas: 2
     template:
       metadata:
         labels:
           app: gcpnfsvolume-demo
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
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=gcpnfsvolume-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```

   The command should print a result similar to the following one:

   ```console
   ...
    Hello from gcpnfsvolume-demo-557dc8cbcb-kwwjt: 19
    Hello from gcpnfsvolume-demo-557dc8cbcb-kwwjt: 20
    File content:
    Hello from gcpnfsvolume-demo-557dc8cbcb-vg5zt: 1
    Hello from gcpnfsvolume-demo-557dc8cbcb-kwwjt: 1
    Hello from gcpnfsvolume-demo-557dc8cbcb-vg5zt: 2
    Hello from gcpnfsvolume-demo-557dc8cbcb-kwwjt: 2
    Hello from gcpnfsvolume-demo-557dc8cbcb-vg5zt: 3
    Hello from gcpnfsvolume-demo-557dc8cbcb-kwwjt: 3
   ...
   ```

   >[!NOTE] The `File content:` contains prints from both workloads. This demonstrates the ReadWriteMany capability of the volume.

## Next Steps

To clean up, follow these steps:

1. Remove the created workloads:

   ```shell
   kubectl delete -n $NAMESPACE deployment gcpnfsvolume-demo
   ```

2. Remove the created configmap:

   ```shell
   kubectl delete -n $NAMESPACE configmap my-script
   ```

3. Remove the created gcpnfsvolume:

   ```shell
   kubectl delete -n $NAMESPACE gcpnfsvolume my-vol
   ```

4. Remove the created default iprange:

   > [!NOTE]
   > If you have other cloud resources using the default IpRange,
   > skip this step, and do not delete the default IpRange.

   ```shell
   kubectl delete iprange default
   ```

5. Remove the created namespace:

   ```shell
   kubectl delete namespace $NAMESPACE
   ```
