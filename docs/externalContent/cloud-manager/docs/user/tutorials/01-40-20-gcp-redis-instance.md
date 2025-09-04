# Using GcpRedisInstance Custom Resources

The Cloud Manager module offers a GcpRedisInstance Custom Resource Definition (CRD). When you apply a GcpRedisInstance custom resource (CR), it creates a Memorystore for Redis instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

### Minimal Setup

This example showcases how to instantiate Redis using only the required fields, connect a Pod to it, and send a PING command.

1. Create a Redis instance. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpRedisInstance
   metadata:
     name: gcpredisinstance-simple-example
   spec:
     redisTier: "S1"
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: gcpredisinstance-simple-example-probe
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
             name: gcpredisinstance-simple-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: gcpredisinstance-simple-example
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: gcpredisinstance-simple-example
   ```

3. Execute into the Pod.

   ```bash
   kubectl exec -i -t gcpredisinstance-simple-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT --tls --cacert /mnt/CaCert.pem PING
   ```

   You should receive `PONG` back from the server.

### Advanced Setup

This example showcases how to instantiate Redis by using most of the spec fields, connect a Pod to it, and send a PING command.

1. Instantiate Redis. The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: GcpRedisInstance
   metadata:
     name: gcpredisinstance-complex-example
   spec:
     redisTier: "P1"
     redisVersion: REDIS_7_2
     authEnabled: true
     redisConfigs:
       maxmemory-policy: volatile-lru
       activedefrag: "yes"
     maintenancePolicy:
       dayOfWeek:
         day: "SATURDAY"
         startTime:
             hours: 15
             minutes: 45
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: gcpredisinstance-complex-example-probe
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
             name: gcpredisinstance-complex-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: gcpredisinstance-complex-example
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: gcpredisinstance-complex-example
       volumeMounts:
       - name: mounted
         mountPath: /mnt
     volumes:
     - name: mounted
       secret:
         secretName: gcpredisinstance-complex-example
   ```

3. Execute into the Pod.

   ```bash
   kubectl exec -i -t gcpredisinstance-complex-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls --cacert /mnt/CaCert.pem PING
   ```

   You should receive `PONG` back from the server.
  