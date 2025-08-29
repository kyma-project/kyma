# API Gateway CRD Proposal

This document describes the proposed API for installing the APIGateway component.

## Proposed CR Structure

```yaml
kind: APIGateway
spec:
  enableKymaGateway: true # Part of the custom resource by default
  defaultHost: "example.com" # Use as default host for API Rules. If not defined and `enableKymaGateway: true`, use the Kyma host. If both fields are false, require a full host in API Gateway
  gateways:
    - namespace: "some-ns" # Required
      name: "gateway1" # Required
      servers:
        - hosts: # Creating  more than one host for the same **host:port** configuration results in a `Warning`
            - host: "goat.example.com"
              certificate: "goat-certificate" # If not defined, generate a Gardener certificate
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided, generate a DNS Entry with Gardener 
          port:
            number: 443
            name: https
            protocol: HTTPS
            TLS: MUTUAL
        - hosts:
            - host: "*.goat.example.com"
            - host: "goat1.example.com"
              dnsProviderSecret: "my-namespace/dns-secret" # If provided, generate a DNS Entry with Gardener 
          port:
            number: 80
            name: http
            protocol: HTTP
          httpsRedirect: true # If `Protocol = HTTPS`, set `Warning`
status:
  state: "Warning"
  description: "Cannot have the same host for two gateways"
  conditions:
  - '[...]' # array of *metav1.Condition

```

## Example Use Cases

### The User Wants to Use API Gateway Without Additional Configuration

The user creates an APIGateway CR with no additional configuration.

```yaml
kind: APIGateway
namespace: kyma-system
name: default
```

By default, APIGateway generates a Certificate and DNSEntry for the default Kyma domain. With this configuration, the user can expose their workloads under the Kyma domain.

### The SAP BTP, Kyma Runtime User Wants to Expose Their Workloads Under a Custom Domain

Prerequisite:
- The DNS secret `dns-secret` exists in the `my-namespace` namespace.

The user configures the CR as follows:

```yaml
kind: APIGateway
namespace: kyma-system
name: default
spec:
  gateways:
    - namespace: "my-namespace"
      name: "default"
      servers:
        - hosts:
            - host: "test.example.com"
              dnsProviderSecret: "my-namespace/dns-secret"
            - host: "test2.example.com"
              dnsProviderSecret: "my-namespace/dns-secret"
```

Because it is a managed Kyma cluster (SKR), a DNSProvider with the provided Secret and a DNSEntry are created. If the user does not configure the port type, Istio Gateway is generated with both HTTP and HTTPS. An additional Gardener Certificate is also created and provided for HTTPS.

The user can now expose their Services under the hosts `test.example.com` and `test2.example.com`.

### The User Wants To Expose Their Mongo Instance

The user configures API Gateway as follows:

```yaml
kind: APIGateway
namespace: kyma-system
name: default
spec:
  gateways:
    - namespace: "my-namespace"
      name: "mongo-gateway"
      servers:
        - hosts:
            - host: "mongo.example.com"
          port:
              number: 2379
              name: mongo
              protocol: MONGO
```

It allows access to Mongo DB under the host `mongo.example.com:2379`.
