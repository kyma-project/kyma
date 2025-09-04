# Using AwsRedisInstance Custom Resources

The Cloud Manager module offers an AwsRedisInstance Custom Resource Definition (CRD). When you apply an AwsRedisInstance custom resource (CR), it creates an ElastiCache for Redis OSS instance that is reachable within your Kubernetes cluster network.

## Prerequisites  <!-- {docsify-ignore} -->

You have the Cloud Manager module added.

## Steps

### Minimal Setup

To instantiate Redis and connect the Pod with only the required fields, use the following setup:

1. Create a Redis instance.

   > [!NOTE]
   > The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsRedisInstance
   metadata:
     name: awsredisinstance-minimal-example
   spec:
     redisTier: S1
   ```

2. Instantiate the redis-cli Pod:

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: awsredisinstance-minimal-example-probe
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
             name: awsredisinstance-minimal-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: awsredisinstance-minimal-example
   ```

3. Execute into the Pod:

   ```bash
   kubectl exec -i -t awsredisinstance-minimal-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Install and update CA certificates:

   ```bash
   apt-get update && \
     apt-get install -y ca-certificates && \
     update-ca-c

5. Run a PING command:

   ```bash
   redis-cli -h $HOST -p $PORT --tls PING
   ```
  
   If your setup was successful, you get `PONG` back from the server.

## Advanced Setup

To specify advanced features, such as Redis version, configuration, and maintenance policy, and set up auth, use the following setup:

1. Instantiate Redis.

   > [!NOTE]
   > The operation may take more than 10 minutes.

   ```yaml
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsRedisInstance
   metadata:
     name: awsredisinstance-advanced-example
   spec:
     redisTier: P1
     engineVersion: "7.0"
     authEnabled: true
     parameters:
       maxmemory-policy: volatile-lru
       activedefrag: "yes"
     preferredMaintenanceWindow: sun:23:00-mon:01:30
     autoMinorVersionUpgrade: true
   ```

2. Instantiate the redis-cli Pod.

   ```yaml
   apiVersion: v1
   kind: Pod
   metadata:
     name: awsredisinstance-advanced-example-probe
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
             name: awsredisinstance-advanced-example
       - name: PORT
         valueFrom:
           secretKeyRef:
             key: port
             name: awsredisinstance-advanced-example
       - name: AUTH_STRING
         valueFrom:
           secretKeyRef:
             key: authString
             name: awsredisinstance-advanced-example
   ```

3. Execute into the Pod.

   ```bash
   kubectl exec -i -t awsredisinstance-advanced-example-probe -c redis-cli -- sh -c "clear; (bash || ash || sh)"
   ```

4. Install and update ca-certificates:

   ```bash
   apt-get update && \
     apt-get install -y ca-certificates && \
     update-ca-certificate
   ```

5. Run a PING command.

   ```bash
   redis-cli -h $HOST -p $PORT -a $AUTH_STRING --tls PING
   ```

   If your setup was successful, you get `PONG` back from the server.
