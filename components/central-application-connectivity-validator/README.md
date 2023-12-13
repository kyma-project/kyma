# Central Application Connectivity Validator

## Overview

Central Application Connectivity Validator validates client certificate subjects in the Compass mode of Kyma.
It proxies the requests to the Eventing Publisher Proxy.

## Usage

Central Application Connectivity Validator has the following parameters:
- **proxyPort** is the port on which the reverse proxy is exposed. The default port is `8081`.
- **externalAPIPort** is the port on which the external API is exposed. The default port is `8080`.
- **eventingPathPrefixV1** is the path prefix for which requests are forwarded to the Eventing Publisher V1 API. The default value is `/v1/events`.
- **eventingPathPrefixV2** is the path prefix for which requests are forwarded to the Eventing Publisher V2 API. The default value is `/v2/events`.
- **eventingPublisherHost** is the host and the port of the Eventing Publisher Proxy. The default value is `events-api:8080`.
- **eventingDestinationPath** is the destination path for the requests coming to the Eventing. The default value is `/`.
- **eventingPathPrefixEvents** is the prefix of paths that is directed to the CloudEvents-based Eventing. The default value is `/events`.
- **appNamePlaceholder**  is the path URL placeholder used for the application name. The default value is `%%APP_NAME%%`.
- **cacheExpirationSeconds** is the expiration time for client IDs stored in cache expressed in seconds. The default value is `90`.
- **cacheCleanupIntervalSeconds** is the clean-up interval controlling how often the client IDs stored in cache are removed. The default value is `15`.
- **syncPeriod** is the time in seconds after which the controller should reconcile the Application resource. The default value is `60 seconds`.

### Application Name Placeholder

If the **appNamePlaceholder** parameter is not empty, it defines a placeholder for the application name in the parameters **eventingPathPrefixV1**, **eventingPathPrefixV2**, and **eventingPathPrefixEvents**. This placeholder is replaced on every proxy request with the value from the certificate Common Name (CN).

### Local Cache Refresh

The application **clientIDs** are read from Application resources and cached locally with the TTL (Time to live) defined by the **cacheExpirationSeconds** parameter.
The cache refresh is performed by the controller during reconciliation in intervals defined by the **syncPeriod**.
To prevent cache entries eviction, the value of the **syncPeriod** should be smaller than that of **cacheExpirationSeconds**.

## Details

The certificate subjects are validated using the `X-Forwarded-Client-Cert` header.
After successful client certificate verification defined in the Istio Gateway, the Envoy Proxy adds the header to the request.
The service to which the header is added must have mutual TLS between Istio sidecar Pods enabled.
This is an example `X-Forwarded-Client-Cert` header:
```bash
Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=,By=spiffe://cluster.local/ns/kyma-system/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account
```

Central Application Connectivity Validator forwards only the requests with the `X-Forwarded-Client-Cert` header that contains **Subject** with the following fields corresponding to the Application custom resource:
- **CommonName** is the name of the Application custom resource.
- **Organization** (optional) is the tenant.
- **OrganizationalUnit** (optional) is the group.

## Development

### Generate Mocks

Prerequisites:

- [Mockery](https://github.com/vektra/mockery) 2.0 or higher

To generate mocks, run:

```sh
go generate ./...
```

When adding a new interface to be mocked or when a mock of an existing interface is not being generated, add the following line directly above the interface declaration:

```
//go:generate mockery --name {INTERFACE_NAME}
```