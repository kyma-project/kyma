# NATS Server

[NATS](https://nats.io) is a simple, secure and performant communications system for digital systems, services and devices. NATS is part of the Cloud Native Computing Foundation ([CNCF](https://cncf.io)). NATS has over [30 client language implementations](https://nats.io/download/), and its server can run on-premise, in the cloud, at the edge, and even on a Raspberry Pi. NATS can secure and simplify design and operation of modern distributed systems.

## TL;DR;

```console
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm install my-nats nats/nats
```

## Configuration

### Server Image

```yaml
nats:
  image: nats:2.1.7-alpine3.11
  pullPolicy: IfNotPresent
```

### Limits

```yaml
nats:
  # The number of connect attempts against discovered routes.
  connectRetries: 30

  # How many seconds should pass before sending a PING
  # to a client that has no activity.
  pingInterval: 

  # Server settings.
  limits:
    maxConnections: 
    maxSubscriptions: 
    maxControlLine: 
    maxPayload: 

    writeDeadline: 
    maxPending: 
    maxPings: 
    lameDuckDuration: 

  # Number of seconds to wait for client connections to end after the pod termination is requested
  terminationGracePeriodSeconds: 60
```

### Logging

*Note*: It is not recommended to enable trace or debug in production since enabling it will significantly degrade performance.

```yaml
nats:
  logging:
    debug: 
    trace: 
    logtime: 
    connectErrorReports: 
    reconnectErrorReports: 
```

### TLS setup for client connections

You can find more on how to setup and trouble shoot TLS connnections at:
https://docs.nats.io/nats-server/configuration/securing_nats/tls

```yaml
nats:
  tls:
    secret:
      name: nats-client-tls
    ca: "ca.crt"
    cert: "tls.crt"
    key: "tls.key"
```

## Clustering

If clustering is enabled, then a 3-node cluster will be setup. More info at:
https://docs.nats.io/nats-server/configuration/clustering#nats-server-clustering

```yaml
cluster:
  enabled: true
  replicas: 3

  tls:
    secret:
      name: nats-server-tls
    ca: "ca.crt"
    cert: "tls.crt"
    key: "tls.key"
```

Example:

```sh
$ helm install nats nats/nats --set cluster.enabled=true
```

## Leafnodes

Leafnode connections to extend a cluster. More info at:
https://docs.nats.io/nats-server/configuration/leafnodes

```yaml
leafnodes:
  enabled: true
  remotes:
    - url: "tls://connect.ngs.global:7422"
      # credentials:
      #   secret:
      #     name: leafnode-creds
      #     key: TA.creds
      # tls:
      #   secret:
      #     name: nats-leafnode-tls
      #   ca: "ca.crt"
      #   cert: "tls.crt"
      #   key: "tls.key"

  #######################
  #                     #
  #  TLS Configuration  #
  #                     #
  #######################
  # 
  #  # You can find more on how to setup and trouble shoot TLS connnections at:
  # 
  #  # https://docs.nats.io/nats-server/configuration/securing_nats/tls
  # 
  tls:
    secret:
      name: nats-client-tls
    ca: "ca.crt"
    cert: "tls.crt"
    key: "tls.key"
```

## Setting up External Access

### Using HostPorts

In case of both external access and advertisements being enabled, an
initializer container will be used to gather the public ips.  This
container will required to have enough RBAC policy to be able to make a
look up of the public ip of the node where it is running.

For example, to setup external access for a cluster and advertise the public ip to clients:

```yaml
nats:
  # Toggle whether to enable external access.
  # This binds a host port for clients, gateways and leafnodes.
  externalAccess: true

  # Toggle to disable client advertisements (connect_urls),
  # in case of running behind a load balancer (which is not recommended)
  # it might be required to disable advertisements.
  advertise: true

  # In case both external access and advertise are enabled
  # then a service account would be required to be able to
  # gather the public ip from a node.
  serviceAccount: "nats-server"
```

Where the service account named `nats-server` has the following RBAC policy for example:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nats-server
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: nats-server
rules:
- apiGroups: [""]
  resources:
  - nodes
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: nats-server-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: nats-server
subjects:
- kind: ServiceAccount
  name: nats-server
  namespace: default
```

The container image of the initializer can be customized via:

```yaml  
bootconfig:
  image: natsio/nats-boot-config:latest
  pullPolicy: IfNotPresent
```

### Using LoadBalancers

In case of using a load balancer for external access, it is recommended to disable no advertise 
so that internal ips from the NATS Servers are not advertised to the clients connecting through
the load balancer.

```yaml
nats:
  image: nats:alpine

cluster:
  enabled: true
  noAdvertise: true

leafnodes:
  enabled: true
  noAdvertise: true

natsbox:
  enabled: true
```

Then could use an L4 enabled load balancer to connect to NATS, for example:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nats-lb
spec:
  type: LoadBalancer
  selector:
    app.kubernetes.io/name: nats
  ports:
    - protocol: TCP
      port: 4222
      targetPort: 4222
      name: nats
    - protocol: TCP
      port: 7422
      targetPort: 7422
      name: leafnodes
    - protocol: TCP
      port: 7522
      targetPort: 7522
      name: gateways
```

## Gateways

A super cluster can be formed by pointing to remote gateways.
You can find more about gateways in the NATS documentation:
https://docs.nats.io/nats-server/configuration/gateways

```yaml
gateway:
  enabled: false
  name: 'default'

  #############################
  #                           #
  #  List of remote gateways  #
  #                           #
  #############################
  # gateways:
  #   - name: other
  #     url: nats://my-gateway-url:7522

  #######################
  #                     #
  #  TLS Configuration  #
  #                     #
  #######################
  # 
  #  # You can find more on how to setup and trouble shoot TLS connnections at:
  # 
  #  # https://docs.nats.io/nats-server/configuration/securing_nats/tls
  #
  # tls:
  #   secret:
  #     name: nats-client-tls
  #   ca: "ca.crt"
  #   cert: "tls.crt"
  #   key: "tls.key"
```

## Auth setup

### Auth with a Memory Resolver

```yaml
auth:
  enabled: true

  # Reference to the Operator JWT.
  operatorjwt:
    configMap:
      name: operator-jwt
      key: KO.jwt

  # Public key of the System Account
  systemAccount:

  resolver:
    ############################
    #                          #
    # Memory resolver settings #
    #                          #
    ##############################
    type: memory

    # 
    # Use a configmap reference which will be mounted
    # into the container.
    # 
    configMap:
      name: nats-accounts
      key: resolver.conf
```

### Auth using an Account Server Resolver

```yaml
auth:
  enabled: true

  # Reference to the Operator JWT.
  operatorjwt:
    configMap:
      name: operator-jwt
      key: KO.jwt

  # Public key of the System Account
  systemAccount:

  resolver:
    ##########################
    #                        #
    #  URL resolver settings #
    #                        #
    ##########################
    type: URL
    url: "http://nats-account-server:9090/jwt/v1/accounts/"
```

## JetStream

### Setting up Memory and File Storage

```yaml
nats:
  image: nats:alpine

  jetstream:
    enabled: true

    memStorage:
      enabled: true
      size: 2Gi

    fileStorage:
      enabled: true
      size: 1Gi
      storageDirectory: /data/
      storageClassName: default
```

### Using with an existing PersistentVolumeClaim

For example, given the following `PersistentVolumeClaim`:

```yaml
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: nats-js-disk
  annotations:
    volume.beta.kubernetes.io/storage-class: "default"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
```

You can start JetStream so that one pod is bounded to it:

```yaml
nats:
  image: nats:alpine

  jetstream:
    enabled: true

    fileStorage:
      enabled: true
      storageDirectory: /data/
      existingClaim: nats-js-disk
      claimStorageSize: 3Gi
```

### Clustering example

```yaml

nats:
  image: nats:alpine

  jetstream:
    enabled: true

    memStorage:
      enabled: true
      size: "2Gi"

    fileStorage:
      enabled: true
      size: "1Gi"
      storageDirectory: /data/
      storageClassName: default

cluster:
  enabled: true
  # Cluster name is required, by default will be release name.
  # name: "nats"
  replicas: 3
```

### Basic Authentication and JetStream

```yaml
nats:
  image: nats:alpine

  jetstream:
    enabled: true

    memStorage:
      enabled: true
      size: "2Gi"

    fileStorage:
      enabled: true
      size: "8Gi"
      storageDirectory: /data/
      storageClassName: gp2

cluster:
  enabled: true
  # Can set a custom cluster name
  # name: "nats"
  replicas: 3

auth:
  enabled: true

  systemAccount: sys

  basic:
    accounts:
      sys:
        users:
        - user: sys
          pass: sys
      js:
        jetstream: true
        users:
        - user: foo
```

### NATS Resolver setup example

As of NATS v2.2, the server now has a built-in NATS resolver of accounts.
The following is an example guide of how to get it configured.

```sh
# Create a working directory to keep the creds.
mkdir nats-creds
cd nats-creds

# This just creates some accounts for you to get started.
curl -fSl https://nats-io.github.io/k8s/setup/nsc-setup.sh | sh
source .nsc.env

# You should have some accounts now, at least the following.
nsc list accounts
+-------------------------------------------------------------------+
|                             Accounts                              |
+--------+----------------------------------------------------------+
| Name   | Public Key                                               |
+--------+----------------------------------------------------------+
| A      | ABJ4OIKBBFCNXZDP25C7EWXCXOVCYYAGBEHFAG7F5XYCOYPHZLNSJYDF |
| B      | ACVRK7GFBRQUCB3NEABGQ7XPNED2BSPT27GOX5QBDYW2NOFMQKK755DJ |
| SYS    | ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N |
+--------+----------------------------------------------------------+

# Now create an account with JetStream support
export account=JS1
nsc add account  --name $account
nsc edit account --name $account --js-disk-storage -1 --js-consumer -1 --js-streams -1
nsc add user -a $account js-user
```

Next, generate the NATS resolver config. This will be used to fill in the values of the YAML in the Helm template.
For example the result of generating this:

```sh
nsc generate config --sys-account SYS --nats-resolver 

# Operator named KO
operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDRlozRlE0WURNTUc1Q1UzU0FUWVlHWUdQUDJaQU1QUzVNRUdNWFdWTUJFWUdIVzc2WEdBIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJLTyIsInN1YiI6Ik9DSVYyUUZKV0lOWlVBVDVUMllKQklSQzNCNkpLTVNaS1FORjVLR1A0TjNLWjRGRkRWQVdZWENMIiwibmF0cyI6eyJ0eXBlIjoib3BlcmF0b3IiLCJ2ZXJzaW9uIjoyfX0.e3gvJ-C1IBznmbUljeT_wbLRl1akv5IGBS3rbxs6mzzTvf3zlqQI4wDKVE8Gvb8qfTX6TIwocClfOqNaN3k3CQ

# System Account named SYS
system_account: ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N

resolver_preload: {
	ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDR0tWVzJGQUszUE5XQTRBWkhHT083UTdZWUVPQkJYNDZaTU1VSFc1TU5QSUFVSFE0RVRRIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJTWVMiLCJzdWIiOiJBREdGSDROWVY1Vjc1U1ZNNURZU1c1QVdPRDdIMk5SVVdBTU82WExaS0lER1VXWUVYQ1pHNUQ2TiIsIm5hdHMiOnsibGltaXRzIjp7InN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsImltcG9ydHMiOi0xLCJleHBvcnRzIjotMSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjotMSwibGVhZiI6LTF9LCJkZWZhdWx0X3Blcm1pc3Npb25zIjp7InB1YiI6e30sInN1YiI6e319LCJ0eXBlIjoiYWNjb3VudCIsInZlcnNpb24iOjJ9fQ.J7g73TEn-ZT13owq4cVWl4l0hZnGK4DJtH2WWOZmGbefcCQ1xsx4cIagKc1cZTCwUpELVAYnSkmPp4LsQOspBg,
}
```

In the YAML would be configured as follows:

```
auth:
  enabled: true

  timeout: "5s"

  resolver:
    type: full

    operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDRlozRlE0WURNTUc1Q1UzU0FUWVlHWUdQUDJaQU1QUzVNRUdNWFdWTUJFWUdIVzc2WEdBIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJLTyIsInN1YiI6Ik9DSVYyUUZKV0lOWlVBVDVUMllKQklSQzNCNkpLTVNaS1FORjVLR1A0TjNLWjRGRkRWQVdZWENMIiwibmF0cyI6eyJ0eXBlIjoib3BlcmF0b3IiLCJ2ZXJzaW9uIjoyfX0.e3gvJ-C1IBznmbUljeT_wbLRl1akv5IGBS3rbxs6mzzTvf3zlqQI4wDKVE8Gvb8qfTX6TIwocClfOqNaN3k3CQ

    systemAccount: ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N

    store:
      dir: "/etc/nats-config/accounts/jwt"
      size: "1Gi"

    resolverPreload:
      ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDR0tWVzJGQUszUE5XQTRBWkhHT083UTdZWUVPQkJYNDZaTU1VSFc1TU5QSUFVSFE0RVRRIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJTWVMiLCJzdWIiOiJBREdGSDROWVY1Vjc1U1ZNNURZU1c1QVdPRDdIMk5SVVdBTU82WExaS0lER1VXWUVYQ1pHNUQ2TiIsIm5hdHMiOnsibGltaXRzIjp7InN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsImltcG9ydHMiOi0xLCJleHBvcnRzIjotMSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjotMSwibGVhZiI6LTF9LCJkZWZhdWx0X3Blcm1pc3Npb25zIjp7InB1YiI6e30sInN1YiI6e319LCJ0eXBlIjoiYWNjb3VudCIsInZlcnNpb24iOjJ9fQ.J7g73TEn-ZT13owq4cVWl4l0hZnGK4DJtH2WWOZmGbefcCQ1xsx4cIagKc1cZTCwUpELVAYnSkmPp4LsQOspBg
```

Now we start the server with the NATS Account Resolver (`auth.resolver.type=full`):

```yaml
nats:
  image: nats:2.6.1-alpine

  logging:
    debug: false
    trace: false

  jetstream:
    enabled: true

    memStorage:
      enabled: true
      size: "2Gi"

    fileStorage:
      enabled: true
      size: "4Gi"
      storageDirectory: /data/
      storageClassName: gp2 # NOTE: AWS setup but customize as needed for your infra.

cluster:
  enabled: true
  # Can set a custom cluster name
  name: "nats"
  replicas: 3

auth:
  enabled: true

  timeout: "5s"

  resolver:
    type: full

    operator: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDRlozRlE0WURNTUc1Q1UzU0FUWVlHWUdQUDJaQU1QUzVNRUdNWFdWTUJFWUdIVzc2WEdBIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJLTyIsInN1YiI6Ik9DSVYyUUZKV0lOWlVBVDVUMllKQklSQzNCNkpLTVNaS1FORjVLR1A0TjNLWjRGRkRWQVdZWENMIiwibmF0cyI6eyJ0eXBlIjoib3BlcmF0b3IiLCJ2ZXJzaW9uIjoyfX0.e3gvJ-C1IBznmbUljeT_wbLRl1akv5IGBS3rbxs6mzzTvf3zlqQI4wDKVE8Gvb8qfTX6TIwocClfOqNaN3k3CQ

    systemAccount: ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N

    store:
      dir: "/etc/nats-config/accounts/jwt"
      size: "1Gi"

    resolverPreload:
      ADGFH4NYV5V75SVM5DYSW5AWOD7H2NRUWAMO6XLZKIDGUWYEXCZG5D6N: eyJ0eXAiOiJKV1QiLCJhbGciOiJlZDI1NTE5LW5rZXkifQ.eyJqdGkiOiJDR0tWVzJGQUszUE5XQTRBWkhHT083UTdZWUVPQkJYNDZaTU1VSFc1TU5QSUFVSFE0RVRRIiwiaWF0IjoxNjMyNzgzMDk2LCJpc3MiOiJPQ0lWMlFGSldJTlpVQVQ1VDJZSkJJUkMzQjZKS01TWktRTkY1S0dQNE4zS1o0RkZEVkFXWVhDTCIsIm5hbWUiOiJTWVMiLCJzdWIiOiJBREdGSDROWVY1Vjc1U1ZNNURZU1c1QVdPRDdIMk5SVVdBTU82WExaS0lER1VXWUVYQ1pHNUQ2TiIsIm5hdHMiOnsibGltaXRzIjp7InN1YnMiOi0xLCJkYXRhIjotMSwicGF5bG9hZCI6LTEsImltcG9ydHMiOi0xLCJleHBvcnRzIjotMSwid2lsZGNhcmRzIjp0cnVlLCJjb25uIjotMSwibGVhZiI6LTF9LCJkZWZhdWx0X3Blcm1pc3Npb25zIjp7InB1YiI6e30sInN1YiI6e319LCJ0eXBlIjoiYWNjb3VudCIsInZlcnNpb24iOjJ9fQ.J7g73TEn-ZT13owq4cVWl4l0hZnGK4DJtH2WWOZmGbefcCQ1xsx4cIagKc1cZTCwUpELVAYnSkmPp4LsQOspBg
```

Finally, using a local port-forward make it possible to establish a connection to one of the servers and upload the accounts.

```sh
nsc push --system-account SYS -u nats://localhost:4222 -A 
[ OK ] push to nats-server "nats://localhost:4222" using system account "SYS":
       [ OK ] push JS1 to nats-server with nats account resolver:
              [ OK ] pushed "JS1" to nats-server nats-0: jwt updated
              [ OK ] pushed "JS1" to nats-server nats-1: jwt updated
              [ OK ] pushed "JS1" to nats-server nats-2: jwt updated
              [ OK ] pushed to a total of 3 nats-server
```

Now you should be able to use JetStream and the NATS based account resolver:

```sh
nats stream ls -s localhost --creds ./nsc/nkeys/creds/KO/JS1/js-user.creds
No Streams defined
```

## Misc

### NATS Box

A lightweight container with NATS and NATS Streaming utilities that is deployed along the cluster to confirm the setup.
You can find the image at: https://github.com/nats-io/nats-box

```yaml
natsbox:
  enabled: true
  image: nats:alpine
  pullPolicy: IfNotPresent

  # credentials:
  #   secret:
  #     name: nats-sys-creds
  #     key: sys.creds
```

### Configuration Reload sidecar

The NATS config reloader image to use:

```yaml
reloader:
  enabled: true
  image: natsio/nats-server-config-reloader:latest
  pullPolicy: IfNotPresent
```

### Prometheus Exporter sidecar

You can toggle whether to start the sidecar that can be used to feed metrics to Prometheus:

```yaml
exporter:
  enabled: true
  image: natsio/prometheus-nats-exporter:latest
  pullPolicy: IfNotPresent
```

### Prometheus operator ServiceMonitor support

You can enable prometheus operator ServiceMonitor:

```yaml
exporter:
  # You have to enable exporter first
  enabled: true
  serviceMonitor:
    enabled: true
    ## Specify the namespace where Prometheus Operator is running
    # namespace: monitoring
    # ...
```

### Pod Customizations

#### Security Context

```yaml
 # Toggle whether to use setup a Pod Security Context
 # ## ref: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/
securityContext:
  fsGroup: 1000
  runAsUser: 1000
  runAsNonRoot: true
```

#### Affinity

<https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity>

`matchExpressions` must be configured according to your setup

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: node.kubernetes.io/purpose
              operator: In
              values:
                - nats
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app
              operator: In
              values:
                - nats
                - stan
        topologyKey: "kubernetes.io/hostname"
```

#### Service topology

[Service topology](https://kubernetes.io/docs/concepts/services-networking/service-topology/) is disabled by default, but can be enabled by setting `topologyKeys`. For example:

```yaml
topologyKeys:
  - "kubernetes.io/hostname"
  - "topology.kubernetes.io/zone"
  - "topology.kubernetes.io/region"
```

#### CPU/Memory Resource Requests/Limits
Sets the pods cpu/memory requests/limits

```yaml
nats:
  resources:
    requests:
      cpu: 2
      memory: 4Gi
    limits:
      cpu: 4
      memory: 6Gi
```

No resources are set by default.

#### Annotations

<https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations>

```yaml
podAnnotations:
  key1 : "value1",
  key2 : "value2"
```

### Name Overides

Can change the name of the resources as needed with:

```yaml
nameOverride: "my-nats"
```

### Image Pull Secrets

```yaml
imagePullSecrets:
- name: myRegistry
```

Adds this to the StatefulSet:

```yaml
spec:
  imagePullSecrets:
    - name: myRegistry
```
