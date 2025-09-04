# Using AzureRedisCluster Custom Resources

The Cloud Manager module offers a AzureRedisCluster Custom Resource Definition (CRD). When you apply a AzureRedisCluster custom resource (CR), it creates a Azure Cache for Redis cluster that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

### Minimal Setup

This example showcases how to instantiate Redis cluster, connect a Pod to it, and send a PING command.

1. Create a Redis cluster. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AzureRedisCluster
   metadata:
     name: azurerediscluster-simple-example
   spec:
     ipRange:
       name: ""
     redisTier: "C3"
     shardCount: "2"
     replicasPerPrimary: "2"
     redisVersion: "6.0"
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
      name: azurerediscluster-simple-example-probe
      labels:
         app: redis-client
   spec:
      containers:
         - name: redis-cli
           image: redis:latest
           resources:
             limits:
               memory: 512Mi
               cpu: "1"
             requests:
               memory: 256Mi
               cpu: "0.2"
           command:
             - "/bin/bash"
             - "-c"
             - "--"
           args:
             - |
               # Install CA certificates
               apt-get update && apt-get install -y ca-certificates && update-ca-certificates;
   
               sleep 86400 & wait
   ```

3. Execute into the Pod.

   ```bash
   kubectl exec -i -t azurerediscluster-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH --tls PING
   ```

   You should receive `PONG` back from the server.
