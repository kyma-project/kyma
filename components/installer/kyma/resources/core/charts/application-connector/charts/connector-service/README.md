```
   ____                            _              
  / ___|___  _ __  _ __   ___  ___| |_ ___  _ __  
 | |   / _ \| '_ \| '_ \ / _ \/ __| __/ _ \| '__|
 | |__| (_) | | | | | | |  __/ (__| || (_) | |    
  \____\___/|_| |_|_| |_|\___|\___|\__\___/|_|    
 / ___|  ___ _ ____   _(_) ___ ___                
 \___ \ / _ \ '__\ \ / / |/ __/ _ \               
  ___) |  __/ |   \ V /| | (_|  __/               
 |____/ \___|_|    \_/ |_|\___\___|                                                                       
```

## Overview
The Connector Service is responsible for generating and sending back client certificates based on Certificate Signing Request (CSR).

## Configuration
The Connector Service has the following parameters, that can be set through the chart:
- **appName** - This is the name of the application used by Kubernetes deployments and services. The default value is `connector-service`.
- **externalAPIPort** - This port exposes the Connector Service API to an external solution. The default port is `8081`.
- **internalAPIPort** - This port exposes the Connector Service within Kubernetes cluster. The default port is `8080`.
- **namespace** - Namespace where Connector Service is deployed. The default Namespace is `kyma-integration`.
- **tokenLength** - Length of registration tokens. The default value is `64`.
- **tokenExpirationMinutes** - Time after which tokens expire and are no longer valid. The default value is `60` minutes.
- **domainName** - Domain name of the cluster, used for generating URL. Default domain name is `.wormhole.cluster.kyma.cx`.
- **certificateServiceHost** - Host at which this service is accessible, used for generating URL. Default host is `cert-service.wormhole.cluster.kyma.cx`.

The Connector Service also uses the following environmental variables for CSR-related information config:
- **COUNTRY** (two-letter-long country code)
- **ORGANIZATION**
- **ORGANIZATIONALUNIT**
- **LOCALITY**
- **PROVINCE**
