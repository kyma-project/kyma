# Central Application Connectivity Validator

## Overview

The Central Application Connectivity Validator validates client certificate subjects.
It proxies the requests to the Event Service and the Application Registry.

The Central Application Connectivity Validator has the following parameters:
- **proxyPort** is the port on which the reverse proxy is exposed. The default port is `8081`.
- **externalAPIPort** is the port on which the external API is exposed. The default port is `8080`.
- **tenant** is the name of the tenant (subject field `OrganizationalUnit`) for which the client certificate should be generated. When empty, the tenant validation is skipped.
- **group** is the name of the group (subject field `Organization`) for which the client certificate should be generated. When empty, the group validation is skipped.
- **eventServicePathPrefixV1** is the path prefix for which requests are forwarded to the Event Service V1 API. The default value is `/{APP_NAME}/v1/events`.
- **eventServicePathPrefixV2** is the path prefix for which requests are forwarded to the Event Service V2 API. The default value is `/{APP_NAME}/v2/events`.
- **eventMeshHost** is the host and the port of the Event Mesh adapter. The default value is `eventing-event-publisher-proxy.kyma-system`.
- **eventMeshDestinationPath** is the destination path for the requests coming to the Event Mesh. The default value is `/publish`.
- **appRegistryPathPrefix** is the path prefix for which requests are forwarded to the Application Registry. The default value is `/{APP_NAME}/v1/metadata`.
- **appRegistryHost** is the host and the port of the Event Service. The default value is `application-registry-external-api:8081`.
- **appNamePlaceholder**  is the path URL placeholder used for the application name. The default value is `%%APP_NAME%%`.
- **cacheExpirationSeconds** is the expiration time for client IDs stored in cache expressed in seconds. The default value is `90`.
- **cacheCleanupIntervalSeconds** is the clean-up interval controlling how often the client IDs stored in cache are removed. The default value is `15`.
- **kubeConfig** is the path to the cluster kubeconfig. Used for running the service outside of the cluster.
- **apiServerURL** is the address of the Kubernetes API server. Overrides any value in the kubeconfig. Used for running the service outside of the cluster.
- **syncPeriod** is the period of time, in seconds, after which the controller should reconcile the Application resource. The default value is `60 seconds`.

### Application name placeholder

If the `appNamePlaceholder` parameter is not empty, it defines a placeholder for the application name in the parameters `eventServicePathPrefixV1`, `eventServicePathPrefixV2`, `eventMeshPathPrefix` and `appRegistryPathPrefix`. This placeholder is replaced on every proxy request
with the value from the certificate Common Name (CN).

### Local cache refresh

The application `clientIDs` are read from Application resources and cached locally with TTL defined by the `cacheExpirationSeconds` parameter.
The cache refresh is performed by the controller during reconciliation in intervals defined by the `syncPeriod`.
To prevent cache entries eviction, the value of the `syncPeriod` should be smaller than that of `cacheExpirationSeconds`.

## Details

The certificate subjects are validated using the `X-Forwarded-Client-Cert` header.
After successful client certificate verification defined in the Istio Gateway, the Envoy Proxy adds the header to the request.
The service to which the header is added must have mutual TLS between Istio sidecar Pods enabled.
This is an example `X-Forwarded-Client-Cert` header:
```
Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account
```

The header contains information about multiple certificates because of the client certificate used in mTLS-secure communication between sidecars of a service.

The Central Application Connectivity Validator forwards only the requests with the `X-Forwarded-Client-Cert` header that contains **Subject** with the following fields corresponding to the Application custom resource:
- **CommonName** is the name of the Application custom resource.
- **Organization** (optional) is the tenant.
- **OrganizationalUnit** (optional) is the group.

## Development

### Generate mocks

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
