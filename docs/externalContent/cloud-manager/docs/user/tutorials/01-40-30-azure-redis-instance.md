# Using AzureRedisInstance Custom Resources

The Cloud Manager module offers a AzureRedisInstance Custom Resource Definition (CRD). When you apply a AzureRedisInstance custom resource (CR), it creates a Azure Cache for Redis instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

### Minimal Setup

This example showcases how to instantiate Redis using only the required fields, connect a Pod to it, and send a PING command.

1. Create a Redis instance. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AzureRedisInstance
   metadata:
     name: azureredisinstance-simple-example
   spec:
     ipRange:
       name: ""
     redisTier: "S1"
     redisVersion: "6.0"
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
      name: azureredisinstance-simple-example-probe
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
   kubectl exec -i -t azureredisinstance-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH --tls PING
   ```

   You should receive `PONG` back from the server.
