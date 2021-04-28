# Central Application Connectivity Validator

## Overview

The Central Application Connectivity Validator validates client certificate subjects.
It proxies the requests to the Event Service and the Application Registry.

The Central Application Connectivity Validator has the following parameters:
- **proxyPort** is the port on which the reverse proxy is exposed. The default port is `8081`.
- **externalAPIPort** is the port on which the external API is exposed. The default port is `8080`.
- **tenant** is the tenant of the Application for which the proxy is deployed. Omitted if empty.
- **group** is the group of the Application for which the proxy is deployed. Omitted if empty.
- **eventServicePathPrefixV1** is the path prefix for which requests are forwarded to the Event Service V1 API. The default value is `/%%APP_NAME%%/v1/events`.
- **eventServicePathPrefixV2** is the path prefix for which requests are forwarded to the Event Service V2 API. The default value is `/%%APP_NAME%%/v2/events`.
- **eventServiceHost** is the host and the port of the Event Service. The default value is `events-api:8080`.
- **eventMeshDestinationPath** is the destination path for the requests coming to the Event Mesh. The default value is `/`.
- **appRegistryPathPrefix** is the path prefix for which requests are forwarded to the Application Registry. The default value is `/%%APP_NAME%%/v1/metadata`.
- **appRegistryHost** is the host and the port of the Event Service. The default value is `application-registry-external-api:8081`.
- **kubeConfig** is the path to a cluster kubeconfig. Used for running the service outside of the cluster.
- **apiServerURL** is the address of the Kubernetes API server. Overrides any value in kubeconfig. Used for running the service outside of the cluster.
- **syncPeriod** is the period of time, in seconds, after which the controller should reconcile the Application resource. The default value is `120 seconds`.
- **appNamePlaceholder**  is the path URL placeholder used for an application name. The default value is `%%APP_NAME%%`.

## Details

The certificate subjects are validated using the `X-Forwarded-Client-Cert` header.
After successful client certificate verification defined in the Istio Gateway, the Envoy Proxy adds the header to the request.
The service to which the header is added must have mutual TLS between Istio sidecar Pods enabled.
This is an example `X-Forwarded-Client-Cert` header:
```
Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account
```

The header contains information about multiple certificates because of the client certificate used in mTLS-secure communication between sidecars of a service.

The Central Application Connectivity Validator forwards only the requests with the `X-Forwarded-Client-Cert` header that contains `Subject` with the following fields corresponding to the Application custom resource:
- **CommonName** is the name of the Application custom resource.
- (Optional) **Organization** is the tenant.
- (Optional) **OrganizationalUnit** is the group.

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
