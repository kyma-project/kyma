# Connector Service

## Overview

The Connector Service is responsible for generating and sending back client certificates based on Certificate Signing Request (CSR).

## Configuration

The Connector Service has the following parameters, that can be set through the chart:
- **appName** - This is the name of the application used by Kubernetes deployments and services. The default value is `connector-service`.
- **externalAPIPort** - This port exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** - This port exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** - Namespace where Connector Service is deployed. The default Namespace is `kyma-integration`.
- **tokenLength** - Length of registration tokens. The default value is `64`.
- **appTokenExpirationMinutes** - Time after which tokens for applications expire and are no longer valid. The default value is `5` minutes.
- **runtimeTokenExpirationMinutes** - Time after which tokens for runtimes expire and are no longer valid. The default value is `10` minutes.
- **caSecretName** - Name of the secret which contains the root Certificate Authority (CA). The default value is `connector-service-app-ca`.
- **requestLogging** - Flag for logging incoming requests. It is set to `False` by default.
- **connectorServiceHost** - Host under which this service is accessible. It is used for generating the URL. The default host is `cert-service.wormhole.cluster.kyma.cx`.
- **appRegistryHost** - Host under which the Application Registry is accessible. The default value is an empty string.
- **eventsHost** - Host under which the Event Service is accessible. The default value is an empty string.
- **appsInfoURL** - URL at which management information for applications is available. If not provided, it bases on `connectorServiceHost`.
- **runtimesInfoURL** - URL at which management information for runtimes is available. If not provided, it bases on `connectorServiceHost`.
- **group** - Group for which certificates are generated. If the chart does not provide the default value, you must specify it in the request header to the token endpoint.
- **tenant** - Tenant for which certificates are generated. If the chart does not provide the default value, you must specify it in the request header to the token endpoint.
- **certificateValidityTime** - Time after which certificates issued by Connector Service will expire. Allowed time units are: `d`, `h`, `m` for days, hours and minutes, only single time unit can be used. Default value is `92d`.
- **central** - Determines whether the Connector Service works in the central mode.

The Connector Service also uses the following environmental variables for CSR-related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**
