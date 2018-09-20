```

   _____                 _             _____      _        _             
  / ____|               (_)           / ____|    | |      | |            
 | (___   ___ _ ____   ___  ___ ___  | |     __ _| |_ __ _| | ___   __ _
  \___ \ / _ \ '__\ \ / / |/ __/ _ \ | |    / _` | __/ _` | |/ _ \ / _` |
  ____) |  __/ |   \ V /| | (_|  __/ | |___| (_| | || (_| | | (_) | (_| |
 |_____/ \___|_|    \_/ |_|\___\___|  \_____\__,_|\__\__,_|_|\___/ \__, |
                                                                    __/ |
                                                                   |___/
```

## Overview

In Kyma, the Service Catalog consists of the following sub-charts:

- `binding-usage-controller` - The sub-chart relates to the Binding Usage Controller which injects ServiceBindings to a given application.
- `catalog` - A Kubernetes Incubator project which provides a Kubernetes-native workflow for integrating with [Open Service Brokers](https://www.openservicebrokerapi.org/). For more information about the project, see the [official documentation](https://github.com/kubernetes-incubator/service-catalog).
- `catalog-service-api` - A GraphQL service which exposes the details about the Kubernetes resources, such as environments, service instances, or deployments. The chart also enables you to create the resources.
- `catalog-service-ui` - An interface which consumes the `catalog-service-api` to display the resources' details to end-users, and list the packages to subscribe to.
- `catalog-ui` - An application with the Service Catalog page view for the Kyma Console. The view enables to list services and provision their instances in the Service Catalog.
- `etcd` - A store for data.
- `service-instances-ui` - An interface which consumes the `catalog-service-api` to display the service instances' details.

## Details

To learn more about the Service Catalog operations and usage, see the [Overview](../../../../docs/service-catalog/docs/001-overview-service-catalog.md) document.
