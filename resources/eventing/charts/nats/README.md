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
