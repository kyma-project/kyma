# Application Connectivity Validator

## Overview

The Application Connectivity Validator is responsible for validation of client certificate subjects.
It is able to proxy the requests to the following services:
- Event Service
- Application Registry

The component is meant to be deployed  per single Application, with the following parameters:
- **proxyPort** - The port on which the reverse proxy is exposed. The default port is `8080`.
- **tenant** - The Tenant of the Application for which the proxy is deployed. Omitted if empty.
- **group** - The Group of the Application for which the proxy is deployed. Omitted if empty.
- **eventServicePathPrefix** - Path prefix for which requests will be forwarded to the Event Service. The default value is `/v1/events`.
- **eventServiceHost** - Host (and port) of the Event Service. The default value is `events-api:8080`.
- **appRegistryPathPrefix** - Path prefix for which requests will be forwarded to the Application Registry. The default value is `/v1/metadata`.
- **appRegistryHost** - Host (and port) of the Event Service. The default value is `application-registry-external-api:8081`.


## Details

The validation is based on `X-Forwarded-Client-Cert` header. 
The header is added to the request by the Envoy Proxy, after successful client certificate validation, which is configured in the Istio Gateway.
For the header to be passed down to the service the mutual TLS between Istio sidecars have to be enabled.

Example `X-Forwarded-Client-Cert` header has the following format:
`Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject="CN=test-application,OU=OrgUnit,O=Organization,L=Waldorf,ST=Waldorf,C=DE";URI=,By=spiffe://cluster.local/ns/kyma-integration/sa/default;Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject="";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account`

Multiple certificate information is due to the client certificate used in mutual TLS communication between sidecars.


The Application Connectivity Validator will forward only requests which contain the `X-Forwarded-Client-Cert` header containing `Subject` with the following fields corresponding to the Application Custom Resource:
- **CommonName** - Application Custom Resource name
- (Optional) **Organization** - Tenant 
- (Optional) **OrganizationalUnit** - Group
