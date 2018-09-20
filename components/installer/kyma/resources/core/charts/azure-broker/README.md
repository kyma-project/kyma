```
                                 ____            _
    /\                          |  _ \          | |
   /  \    _____   _ _ __ ___   | |_) |_ __ ___ | | _____ _ __
  / /\ \  |_  / | | | '__/ _ \  |  _ <| '__/ _ \| |/ / _ \ '__|
 / ____ \  / /| |_| | | |  __/  | |_) | | | (_) |   <  __/ |
/_/    \_\/___|\__,_|_|  \___|  |____/|_|  \___/|_|\_\___|_|

```

## Overview

The [Azure Broker](https://github.com/Azure/open-service-broker-azure) is an open source, [Open Service Broker](https://www.openservicebrokerapi.org/)-compatible
API server that provisions managed services in the Microsoft Azure public cloud.
This chart is based on the [Azure Open Service Broker](https://github.com/Azure/open-service-broker-azure/tree/master/contrib/k8s/charts/open-service-broker-azure) chart and runs in Kyma.


## Details

The [azure-broker-basic-auth](templates/azure-broker-basic-auth.yaml) file contains the user name and password for basic authentication used by the Service Catalog.
The [azure-broker-redis](templates/azure-broker-redis.yaml) file contains Redis specific data.

For more information about the Service Brokers, see [this](../../../../docs/service-brokers/docs) repository.
