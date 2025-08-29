# Geoblocking

## Status

Accepted

## Context

Geoblocking is a feature that allows blocking incoming traffic from certain IP addresses exclusive to certain countries or regions.
It can be utilized against anonymous users and system-to-system network communication, where both are identified only by its source IP address.
Many companies use geoblocking to protect their services from unwanted traffic, such as DDoS attacks, or to comply with legal regulations.
Therefore, creating this kind of feature is convenient for the user because it eliminates the need for them to implement it themselves.

## Decision

Istio allows delegating access control to an external authorization service, which may be just a regular Kubernetes service. This seems to be the simplest way to integrate the geoblocking check of incoming connections.

### ip-auth Service

For that purpose, a new service, 'ip-auth', is introduced. Its main responsibility is to fulfill the [external authorizer](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/) contract by:
- listening to the connections from the Envoy proxy
- deciding whether to allow or disallow an incoming connection based on headers and the current list of IP ranges
- responding with HTTP `200` to allow an incoming connection or with HTTP `403` to disallow it

Here's the high-level overview:

![IP Auth](../../assets/geoblocking.drawio.svg)

### Modes of Operation

The IP Auth service offers two modes of operation:
1. It uses an IP range allow/block list populated by the customer.
2. It uses a SAP internal service (only for SAP internal customers).

#### IP range Allow/Block List Populated by the Customer

In this mode, the list of blocked IP ranges is read from a ConfigMap and stored in the ip-auth application's memory. The end-user may update the list of IP ranges at any time, so the ip-auth application is obliged to refresh it regularly. 

See an example of an IP range allow/block list:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: ip-range-list
data:
  allow: |
    - 192.0.2.10/32
    - 2001:0002:6c::430
  block: |
    - 192.0.2.0/24
    - 198.51.100.0/24
    - 203.0.113.0/24
    - 2001:0002::/48
```

The allow list should take precedence over the block list, which may be useful in allowing a narrower range within a blocked broader range. The lists may contain both IPv4 and IPv6 ranges.

![Static list](../../assets/geoblocking-custom-list.drawio.svg)

#### Usage of SAP Internal Service

In this mode, an SAP internal service receives the list of blocked IP ranges. In order to connect to it, ip-auth requires a secret with SAP internal service credentials. The list of blocked IP ranges is then in application memory and additionally in a ConfigMap, which works as a persistent cache. This approach limits the number of list downloads and makes the whole solution more reliable if the SAP internal service is not available. The list of IP ranges should be refreshed once per hour.

Additionally, the ip-auth Service uses the SAP internal service to report the following events:
- policy list consumption (success, failure, unchanged)
- access decision (allow, deny)

The ip-auth Service should store the IP range allow/block list in a Config Map for caching/fallback purposes. In this case, the ConfigMap may contain more attributes:
- **version** - used in events
- **etag** - used when checking for updates
- **lastUpdateCheckTime** - the time when the list update has been performed successfully
- **lastUpdateTime** - the time when the list has been updated

The above Config Map should only be used by the ip-auth Service for caching/fallback purpose. It can't be treated as an official contract. It may be replaced with some other solution at any time without further notice.

See an example of an IP range allow/block list cache:

```
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ip-range-list-cache
data:
  version: ip-list-20240912-2200
  etag: PeJwSe7ncNQchtGbkCPJ
  lastUpdateCheckTime: "2024-09-17T15:28:50Z"
  lastUpdateTime: "2024-09-17T14:28:50Z"
  allow: |
    - 192.0.2.10/32
    - 2001:0002:6c::430
  block: |
    - 192.0.2.0/24
    - 198.51.100.0/24
    - 203.0.113.0/24
    - 2001:0002::/48
```

![SAP internal service](../../assets/geoblocking-SAP-service.drawio.svg)

### GeoBlocking CR and Controller

In order to ensure reliability and configurability, a new GeoBlocking custom resource and a new GeoBlocking Controller is introduced. The controller is responsible for:
- managing the ip-auth Service's Deployment
- managing authorization policy that plugs ip-auth authorizer to all incoming requests
- performing configuration checks (like external traffic policy)
- reporting geoblocking state

There should be only one GeoBlocking resource. The controller should set the additional GeoBlocking CR in the `Error` state, similarly to [APIGateway CR](https://github.com/kyma-project/api-gateway/blob/05ed57f3299f42c2565ed1b28b84dee808a5213b/controllers/operator/apigateway_controller.go#L100).

![GeoBlocking Controller](../../assets/geoblocking-cr-controller.drawio.svg)

#### CR Examples

##### Usage of SAP Internal Service

```yaml
apiVersion: geoblocking.kyma-project.io/v1alpha1
kind: GeoBlocking
metadata:
  name: geoblocking
  namespace: kyma-system
spec:
  gbService:
    secret: "gb-secret"
    tokenUrl: https://oauth.example.com/oauth2/token
    ipRange:
      url: https://lists.example.com/lists/some-list
      refreshInSeconds: 3600
    events:
      url: https://events.example.com/events
      lobId: "kyma"
      productId: "some-product"
      systemId: "some-system"
```

##### IP Range Allow/Block List Populated by the Customer
```yaml
apiVersion: geoblocking.kyma-project.io/v1alpha1
kind: GeoBlocking
metadata:
    name: geoblocking
    namespace: kyma-system
spec:
  ipRange:
    configMap: "my-own-ip-range-list"
```

### Technical Details

#### Policy Download Optimization

In order to reduce unnecessary updates to the list of blocked IP ranges, the ip-auth Service should use the ETag mechanism, which is supported by the SAP internal service.

The local copy (stored in a Config Map) should be used to reduce the amount of connections to the SAP internal service. The idea is that only one Pod should contact the SAP internal service (if the time interval defined in **refreshInterval** has passed). So the algorithm may use the following rules:

- if the local copy contains a newer version of the policy list (lastUpdateTime attribute) -> it means that another pod has already downloaded it from the SAP internal service, so load the policy list from the local copy
- if the update check has been recently performed (now - **lastUpdateCheckTime** < **refreshInterval**) -> don't check it again
- if the update check hasn't been recently performed (now - **lastUpdateCheckTime** >= **refreshInterval**) -> check whether a newer version of the policy list is available in the SAP internal service
- if there is a problem with checking for an update -> log a warning
- if there is a newer version available in the SAP internal service -> download, apply it, store it in the local copy, set both **lastUpdateCheckTime** and **lastUpdateTime** attributes to the current time
- if there is a problem with storing the newer version of the policy list it the local copy -> log a warning (but apply a new policy version anyway)
- if there is no newer version available in the SAP internal service -> set **lastUpdateCheckTime** attribute to the current time, so other pods may skip this check (until the time defined in **refreshInterval** passes again)

Pods should slightly randomize an update check time to benefit from the above optimization and minimize the risk of performing the check by multiple Pods at the same time.

#### Quick IP Check

The list of blocked IP ranges may be big (thousands of entries), so analyzing whether an IP address matches any IP range in the list may be time consuming. The 'linear search' through all IP ranges would probably be too slow.

Because the list of IP ranges changes rarely (compared to the frequency of incoming connections), it is recommended to build an efficient data structure that supports quick comparison of IP addresses. The good candidate is a [radix tree](https://en.wikipedia.org/wiki/Radix_tree) or something like [TSS](https://citeseerx.ist.psu.edu/document?repid=rep1&type=pdf&doi=3aa1dd14e3d1c20d1f09b0ce0d4b4dd7b1190885). This can be decided during the implementation phase and would probably require some tests (benchmark, etc.).

#### Events Retention

To prevent unnecessary delays, the access events should be sent asynchronously:
- After making an access decision, ip-auth should generate an event and store it in a memory queue.
- A separate thread should retrieve events from the queue and send them to the SAP internal service.

This approach may cause issues if the SAP internal service responds slowly or does not respond at all, because events would be consuming memory. System availability is a key factor here, so it is acceptable to drop events in order to ensure stability and prevent out-of-memory issues.

Events should have a retry number so the application can drop them if the maximum number of retries is exceeded or if the queue is full. All such cases should be properly logged.

There won't be any persistent storage for events, so unsent events would be lost in case of ip-auth crash (and SAP internal service instability).

#### Graceful Shutdown/Probes

It would be good to ensure some graceful shutdown mechanism so the ip-auth Pods may send events that are in the queue in case of a controlled shutdown. A good example of such a situation is autoscaling or a rolling update.

IP-auth should also have a readiness probe that cooperates with graceful shutdown mode, so it won't receive new requests if the Pod is being shut down.

#### Headers Used in the Check

The ip-auth Service should take the following HTTP headers into consideration:
1. **x-envoy-external-address** - contains a single trusted IP address
2. **x-forwarded-for** - contains an appendable list of IP addresses modified by each compliant proxy

IP-auth should block the connection request if any IP address in any of the above headers belongs to any IP range in the block list (unless they are also in the allow list).

#### ip-auth Deployment Settings

The main responsibility of the controller is to reconcile the ip-auth Deployment. The Deployment details should be hardcoded for simplicity, similarly to the [oathkeeper Deployment](https://github.com/kyma-project/api-gateway/blob/fdef70be4ca9a9319ae4c547f4f6dad8c73b9846/internal/reconciliations/oathkeeper/deployment.yaml).

This includes at least the following details:
- Deployment:
  - replicas / scaling
  - update strategy
  - Docker image
  - resources (limits, requests)
  - probes (liveness, readiness)
  - exposed ports
  - security context (user, etc.)
  - service account
  - configmap mount (ip allow / block list)
  - secret mount (secrets for SAP internal service)
  - priority class
- Service:
  - selectors
  - ports

#### Deployment Stability

The ip-auth Deployment stability is quite critical from system availability point of view. Thus, it requires at least the following:
- horizontal autoscaling
- rolling updates (so there are always some pods available during upgrade)
- readiness probe (container won't receive new requests if the IP range list is not available yet)
- liveness probe (non-working container is restarted)
- priority class (pod is scheduled before other less important pods)

#### Authorization Policy Settings

For simplicity reasons, we assume that geoblocking would be applied globally for the whole application.

This means that it applies to the whole Istio Ingress Gateway, and there is no need to expose any additional configuration in this scope.

For now, we assume that the feature works with default Istio Ingress Gateway installed by the Istio module.

#### Namespaces

Geoblocking resources would be placed in the following namespaces:

| Resource                                    | Created By                     | Namespace                            |
| ------------------------------------------- | ------------------------------ | ------------------------------------ |
| GeoBlocking Operator (APIGateway Operator)  | Power user / Lifecycle Manager | `kyma-system`                        |
| GeoBlocking CR                              | Power user                     | `kyma-system`                        |
| ip-auth Deployment                          | GeoBlocking Operator           | `kyma-system`                        |
| ip-auth Service                             | GeoBlocking Operator           | `kyma-system`                        |
| ip-auth Secret                              | Power user                     | `kyma-system`                        |
| IP ranges ConfigMap with input              | ip-auth or external customer   | `kyma-system`                        |
| IP ranges ConfigMap with cache              | ip-auth or external customer   | `kyma-system`                        |
| AuthorizationPolicy                         | GeoBlocking Operator           | `istio-system`                       |

![Geoblocking namespaces](../../assets/geoblocking-namespaces.drawio.svg)

Consequently, two separate Service Accounts are required, so:
- GeoBlocking Operator is able to read CR, create/write ip-auth Deployment, ip-auth Service, check if the ip-auth secret exists, check if ConfigMap with IP ranges exists, create/write Authorization Policy in the `istio-system` namespace, read the Istio configuration
- ip-auth is able to read ip-auth Secret or have it mounted, read the ConfigMap with IP ranges, create/write the ConfigMap with IP ranges cache

## Considered Alternative Architectural Approaches

### Workload resources and event queue parameters

Events are stored in a queue, so there is a correlation between:
- queue size
- event TTL
- ip-auth Pod memory resources

These parameters may be:
- hardcoded
- calculated based on current memory setup
- configurable in the custom resource

Configuration could look like:
```yaml
spec:
  gbService:
    events:
      queue:
        size: 100
        ttl: 360
  deployment: 
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"
```

Decision: It doesn't make sense to expose so many technical parameters, especially that there would be no easy way to determine the above values. It is better to hardcode these parameters with some reasonable values (for now), potentially providing multiple variants (evaluation and production profiles).

### Configurability of the Rules in AuthorizationPolicy

Istio allows to specify a lot of conditions that are used to link a request with a given AuthorizationPolicy. See the [Istio documentation](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule)

We can either:
- expose them and make it configurable in the CR
- hide it and don't allow such configurability

Decision: It doesn't make sense to expose so many detailed parameters, because the idea behind geoblocking is to be compliant with regulations, which are usually very global and apply to the whole company, product, or application.

### IP Range Allow/Block List Fallback/Caching

The list of IP ranges is processed by each ip-auth container, stored in its memory, and regularly refreshed using the SAP internal service. If the SAP internal service doesn't work, then the ip-auth containers may work as before and the list is not refreshed.

However, the problem is when the ip-auth container needs to be restarted when the SAP internal service is not available, so the ip-auth container can't download the list of IP ranges. This may be the case during an upgrade of ip-auth or in the event of runtime problems (for example, out of memory). It is not acceptable to allow the traffic in such a case. This means that the list of IP ranges should be stored in a way that allows it to persist through the Pods' restarts. This requirement is referred to as 'fallback'.

Another aspect is the reduction of the number of connections made to the SAP internal service. Multiple ip-auth Pods independently asking SAP internal service for policy updates would generate unnecessary load and it should be avoided if possible. This requirement is referred to as 'cache'.

The following options have been considered:

| Option             | Advantages                                              | Disadvantages                                           |
| ------------------ | ------------------------------------------------------- | ------------------------------------------------------- |
| Allow all fallback | Very simple                                             | Not acceptable from security perspective                |
|                    |                                                         |                                                         |
| No fallback        | Very simple                                             | Downtime if the SAP internal service doesn't work       |
|                    |                                                         | Every container causes load on the SAP internal service |
|                    |                                                         |                                                         |
| Caching service    | Extensibility (may cache more things in future)         | Complex - another component to take care                |
|                    | Works well with restart / upgrade / scaling             |                                                         |
|                    | No downtime if the SAP internal service doesn't work    |                                                         |
|                    | Less load on the SAP internal service                   |                                                         |
|                    |                                                         |                                                         |
| Config map         | Still simple                                            | Configmap capacity allows max ~50000 IP ranges          |
|                    | Works well with restart / upgrade / scaling             | IP range list integrity not guaranteed                  |
|                    | No downtime if the SAP internal service doesn't work    |                                                         |
|                    | Less load on the SAP internal service                   |                                                         |
|                    |                                                         |                                                         |
| Ephemeral volume   | Still simple                                            | Every container causes load on geoblocking service      |
|                    | No downtime if the SAP internal service doesn't work    | Doesn't help in case of restart / upgrade / scaling     |
|                    |                                                         |                                                         |
| Persistent volume  | No downtime if the SAP internal service doesn't work    | Depends on the cloud infrastructure (ReadWriteMany)     |
|                    | Less load on the SAP internal service                   |                                                         |
|                    | Works well with restart / upgrade / scaling             |                                                         |

Decision: Let's use a ConfigMap as a fallback and a cache for the IP range allow/block list. However, let's split it technically from the ConfigMap that contains IP ranges provided by the end-user, so:
- The ConfigMap containing the custom IP range allow/block list is configured by the user. It becomes a contract.
- The ConfigMap containing the IP range allow/block list downloaded from the SAP internal service is a Kyma internal resource, and the implementation may change at any time, so no other module should use it.

Consequence: The ConfigMap's capacity may be exceeded in the future, which may require immediate attention. It would be good to observe the number of entries in the policy list to be able to react to any potential capacity issues early on. Another consequence is that power users with `kyma-system` namespace access would be able to modify the IP ranges cache, which is not intended, but it is not possible to additionally protect it.

### ip-auth Deployment's Configurability

GeoBlocking Operator's main functionality is to deploy the ip-auth Service. The Deployment requires multiple parameters (like a Docker image and resources). These parameters may be either configurable in the custom resource or just hardcoded.

Decision: Hardcode Deployment details in the API Gateway module.

Consequence: It won't be possible to change the Deployment details. 

### Propagation of ip-auth Service Issues

GeoBlocking Controller is responsible for the reconciliation of the GeoBlocking custom resource, in particular: creating the Deployment of the ip-auth Service, creating AuthorizationPolicy, performing configuration checks, etc. It may report issues related to the above actions in the GeoBlocking custom resource status.

However, some issues may be visible only in the ip-auth Service. A good example is a problem with the communication with the SAP internal service (like a wrong secret).

There is no easy way to propagate such issues to the GeoBlocking CR. This would require a dedicated mechanism, like a 'health' endpoint exposed by ip-auth containers, so the GeoBlocking Controller can analyze them.

Decision: Don't implement additional mechanisms and follow the 'standard' approach, so the controller reports issues in the CR only from its own layer, while the ip-auth Service reports issues in its log or via liveness/readiness probes (like every other workload).

Consequence: The GeoBlocking CR won't be responsible for presenting geoblocking 'health'. Support or developers must be aware that they should also inspect ip-auth logs in case of issues.

### Namespaces

Geoblocking mechanism touches multiple resources in multiple areas:
- Istio configuration
- Ingress Gateway
- Gateways in general
- Authorization Policy
- Authorizer workload (Service, Deployment)
- Configuration of the SAP internal service
- Custom IP range allow/block list
- IP range allow/block list cache

It is a mix-in of different scopes - those managed by Kyma, by Istio, and by the end-users.

There are multiple factors that need to be taken into consideration:
- The operator needs more permissions to access multiple namespaces.
- Kyma resources (workloads) should run in the `kyma-system` namespace.
- The `kyma-system` namespace should not be used by the end-users.
- Ingress Gateway is created by default in the `istio-system` namespace.
- Users may influence Istio (via Gateway API) to create Ingress Gateway in a different namespace.
- AuthorizationPolicy references Ingress Gateway using a selector, so they should be in the same namespace.
- For now, the GeoBlocking CR is supposed to be a singleton.

There are multiple possible approaches:
- The GeoBlocking CR is close to the Gateway. By default, in the `istio-system` namespace.
- The GeoBlocking CR is in the end-user's (workload) namespace.
- The GeoBlocking CR is in `kyma-namespace` (close to ip-auth Deployment).
- ip-auth Deployment is run in the user's namespace (allows multiple user configurations).
- ip-auth Deployment is in the `kyma-system` namespace along with other Kyma modules.
- All resources are in the `istio-system` namespace. This approach avoids using multiple namespaces, but it also disrupts the separation of Istio.

Decision: We can't avoid working with multiple namespaces. Taking into consideration the fact that geoblocking (being a compliance requirement) is rather an application-global functionality, it makes sense to have a single GeoBlocking CR (implying a single configuration). Thus, it seems that the optimal solution is to keep all resources in the `kyma-system` namespace, except for the AuthorizationPolicy resource that references Ingress Gateway, which, by default, is located in the `istio-system`.

Consequences: The power user would have to configure geoblocking in the `kyma-system` namespace by creating the CR and providing a Secret.

### Istio ip-auth Authorizer Declaration

The ip-auth Service must be declared as a custom authorizer in the Istio configuration.

There are several possibilities to do it:
- GeoBlocking Controller may modify Istio configuration. This breaks separation.
- GeoBlocking Controller may modify the Istio CR. This also breaks separation.
- The end-user may modify the Istio CR.
- The Istio module may always configure the Istio CR.

Decision: The end users must declare the ip-auth Service in the Istio CR because they need to modify the CR anyway (for example, configure the external traffic policy). GeoBlocking Operator should check whether it is properly configured by verifying the Authorizer's declaration, External Traffic Policy, etc.

Consequence: Geoblocking won't work OOTB after applying the GeoBlocking CR.
