---
title: Failure to build Functions
type: Troubleshooting
---

In its default configuration, Serverless uses persistent volumes as the internal registry to store Docker images for Functions. The default storage size of such a volume is 20 GB. When this storage becomes full, you will have issues with building your Functions. As a workaround, increase the default capacity up to a maximum of 100 GB by editing the `serverless-docker-registry` PersistentVolumeClaim (PVC) object on your cluster.

Follow these steps:

1. Edit the `serverless-docker-registry` PVC:

  ```bash
  kubectl edit pvc -n kyma-system serverless-docker-registry
  ```

2.  Change the value of **spec.resources.requests.storage** to higher, such as 30 GB, to increase the PVC capacity:

  ```yaml
  ...
  spec:
    resources:
      requests:
        storage: 30Gi
  ```

3. Save the changes and wait for a few minutes. Use this command to check if the **CAPACITY** of the `serverless-docker-registry` PVC has changed as expected:

  ```bash
  kubectl get pvc serverless-docker-registry -n kyma-system
  ```

  You will get this result:

  ```bash
  NAME                        STATUS   VOLUME                                    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
  serverless-docker-registry  Bound    pvc-a69b96hc-ahbc-k85d-0gh6-19gkcr4yns4k  30Gi       RWO            standard       23d
  ```

If the value of the storage does not change, restart the Pod to which this PVC is bound to finish the volume resize.

To do this, follow these steps:

1. List all available Pods in the `kyma-system` Namespace:

  ```bash
  kubectl get pods -n kyma-system
  ```

2. Search for the Pod with the `serverless-docker-registry-{UNIQUE_ID}` name and delete it. See the example below:

  ```bash
  kubectl delete pod serverless-docker-registry-6869bd57dc-9tqxp -n kyma-system
  ```

  >**CAUTION:** Do not remove Pods named `serverless-docker-registry-self-signed-cert-{UNIQUE_ID}`.

3. Search for the `serverless-docker-registry` PVC again to check that the capacity was resized:

  ```bash
  kubectl get pvc serverless-docker-registry -n kyma-system
  ```
