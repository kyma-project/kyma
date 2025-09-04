# Using GcpRedisCluster Custom Resources

The Cloud Manager module offers a GcpRedisCluster Custom Resource Definition (CRD). When you apply a GcpRedisCluster custom resource (CR), it creates a Memorystore for Redis Cluster instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps


This example showcases how to instantiate a Redis cluster, connect a Pod to it, and send a PING command.

1. Create a Redis cluster. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpRedisCluster
   metadata:
     name: gcprediscluster-simple-example
   spec:
     redisTier: "C1"
     shardCount: 3
     replicasPerShard: 2
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: gcprediscluster-simple-example-probe
   spec:
     containers:
     - name: redis-cli
       image: redis:latest
       command: ["/bin/sleep"]
       args: ["999999999999"]
       env:
       - name: HOST
         valueFrom:
           secretKeyRef:
             key: host
             name: gcprediscluster-simple-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: gcprediscluster-simple-example
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: gcprediscluster-simple-example
   ```

3. Execute into the Pod.

   ```bash
   kubectl exec -i -t gcprediscluster-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT --tls --cacert /mnt/CaCert.pem -c PING
   ```

   You should receive `PONG` back from the server.

