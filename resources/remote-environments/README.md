```
  _____                      _         ______            _                                      _       
 |  __ \                    | |       |  ____|          (_)                                    | |      
 | |__) |___ _ __ ___   ___ | |_ ___  | |__   _ ____   ___ _ __ ___  _ __  _ __ ___   ___ _ __ | |_ ___
 |  _  // _ \ '_ ` _ \ / _ \| __/ _ \ |  __| | '_ \ \ / / | '__/ _ \| '_ \| '_ ` _ \ / _ \ '_ \| __/ __|
 | | \ \  __/ | | | | | (_) | ||  __/ | |____| | | \ V /| | | | (_) | | | | | | | | |  __/ | | | |_\__ \
 |_|  \_\___|_| |_| |_|\___/ \__\___| |______|_| |_|\_/ |_|_|  \___/|_| |_|_| |_| |_|\___|_| |_|\__|___/                                                                                                        
```

## Overview

A Remote Environment (RE) is a representation of an external solution connected to Kyma. The traffic, connection, security, and events of REs are managed by the Application Connector - a proprietary implementation that consists of four services.
Read the [Application Connector documentation](../../docs/application-connector/docs/001-overview-application-connector.md) for more details regarding the implementation.

## Details

This directory contains the Remote Environment Helm chart that creates a new RE and installs the Gateway Service and all of the Ingresses required to access Application Connector services in the context of the created RE. A single RE allows to connect a single external solution to Kyma.

### Customize the Gateway Service installation

Edit the [`values`](./values.yaml) file to customize the installation of the Gateway Service.
You can adjust these parameters:

- **proxyPort** - This port proxies calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the Gateway API to an external solution. The default port is `8081`.
- **eventsTargetURL** - The URL to proxy incoming events. The default URL is `http://localhost:9000`.
- **remoteEnvironment** - The RE to read and write information about the services. The default RE is `default-ec`.
- **namespace** - The Namespace to which you deploy the Gateway. The default Namespace is `kyma-system`.
- **requestTimeout** - A time-out for requests sent through the Gateway. Provide it in seconds. The default time-out is `1`.
- **skipVerify** - The flag to skip the verification of certificates for the proxy targets. The default value is `false`.

Additionally, you can adjust the parameters used in the communication with the Event Service:
- **sourceEnvironment** - The Event source environment name.
- **sourceType** - The Event source type.
- **sourceNamespace** - The organization that publishes the Event.
