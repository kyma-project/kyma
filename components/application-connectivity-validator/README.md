# Application Connectivity Validator

## Overview

The Application Connectivity Validator validates client certificate subjects.
It proxies the requests to the Event Service and the Application Registry.
- Event Service
- Application Registry

A single instance of the component is deployed for an Application and uses these parameters:
- **proxyPort** - The port on which the reverse proxy is exposed. The default port is `8080`.
- **tenant** - The tenant of the Application for which the proxy is deployed. Omitted if empty.
- **group** - The group of the Application for which the proxy is deployed. Omitted if empty.
- **eventServicePathPrefix** - Path prefix for which requests will be forwarded to the Event Service. The default value is `/v1/events`.
- **eventServiceHost** - Host and port of the Event Service. The default value is `events-api:8080`.
- **appRegistryPathPrefix** - Path prefix for which requests will be forwarded to the Application Registry. The default value is `/v1/metadata`.
- **appRegistryHost** - Host and port of the Event Service. The default value is `application-registry-external-api:8081`.


## Details

The certificate subjects are validated using the `X-Forwarded-Client-Cert` header. 
The header is added to the request by the Envoy Proxy, after successful client certificate validation, which is configured in the Istio Gateway.
For the header to be passed down to the service the mutual TLS between Istio sidecars have to be enabled.

This is an example `X-Forwarded-Client-Cert` header:
`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`

Multiple certificate information is due to the client certificate used in mutual TLS communication between sidecars.


The Application Connectivity Validator forwards only the requests that contain the `X-Forwarded-Client-Cert` header that contains `Subject` with the following fields corresponding to the Application custom resource:
- **CommonName** - Application custom resource name
- (Optional) **Organization** - Tenant 
- (Optional) **OrganizationalUnit** - Group
