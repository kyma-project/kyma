```
                       _ _           _   _                _____                            _
     /\               | (_)         | | (_)              / ____|                          | |
    /  \   _ __  _ __ | |_  ___ __ _| |_ _  ___  _ __   | |     ___  _ __  _ __   ___  ___| |_ ___  _ __
   / /\ \ | '_ \| '_ \| | |/ __/ _` | __| |/ _ \| '_ \  | |    / _ \| '_ \| '_ \ / _ \/ __| __/ _ \| '__|
  / ____ \| |_) | |_) | | | (_| (_| | |_| | (_) | | | | | |___| (_) | | | | | | |  __/ (__| || (_) | |
 /_/    \_\ .__/| .__/|_|_|\___\__,_|\__|_|\___/|_| |_|  \_____\___/|_| |_|_| |_|\___|\___|\__\___/|_|
          | |   | |
          |_|   |_|
```

## Overview

An Application Connector connects an external solution to Kyma.

## Details

An Application Connector contains a Remote Environment Custom Resource Definition (CRD) and the Gateway service, exposed over the Ingress-Nginx controller, which handles the connection between Kyma and the external solution.

### Remote Environment CRD

The CRD contains the information about a given Remote Environment, as well as the connected external solution and the Event Bus configuration. For more information, see the [Remote Environment definition](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/014-details-remote-environment.md) document.

### Gateway

Gateway is a Kyma core component that proxies events and HTTP requests to and from Kyma.
To define ports, adjust the `values.yaml` file.

The Gateway has the following parameters:

- **proxyPort** - This port proxies calls from services and lambdas to an external solution. The default port is `8080`.
- **externalAPIPort** - This port exposes the Gateway API to an external solution. The default port is `8081`.
- **eventsTargetURL** - The URL to proxy incoming events. The default URL is `http://localhost:9000`.
- **remoteEnvironment** - The Remote Environment to read and write information about the services. The default Remote Environment is `default-ec`.
- **namespace** - The Namespace to which you deploy the Gateway. The default Namespace is `kyma-system`.
- **requestTimeout** - A time-out for requests sent through the Gateway. Provide it in seconds. The default time-out is `1`.
- **skipVerify** - The flag to skip the verification of certificates for the proxy targets. The default value is `false`.

The Gateway has the parameters for the Event API which correspond to the fields in the [Remote Environment](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/014-details-remote-environment.md):

- **sourceEnvironment** - The Event source environment name.
- **sourceType** - The Event source type.
- **sourceNamespace** - The organization that publishes the Event.

### Installation

By default, Kyma comes with these Application Connector installed in the `kyma-integration` Namespace:

- ec-default
- hmc-default

#### Install an Application Connector manually

Use the Helm chart to install an additional Application Connector.

``` bash
helm install --name remote-environment-name --set deployment.args.sourceType=commerce --set global.isLocalEnv=false --namespace kyma-integration ./resources/remote-environments
```

For installations on Minikube, provide the NodePort as shown in this example:

``` bash
helm install --name remote-environment-name --set deployment.args.sourceType=commerce --set global.isLocalEnv=true --set service.externalapi.nodePort=32001 --namespace kyma-integration ./resources/remote-environments
```

The user can override the following parameters:

- **sourceEnvironment**
- **sourceType**
- **sourceNamespace**
