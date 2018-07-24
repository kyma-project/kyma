```
   _____ _           _                                                  _     _ _            
  / ____| |         | |                                                (_)   (_) |           
 | |    | |_   _ ___| |_ ___ _ __   _ __  _ __ ___ _ __ ___  __ _ _   _ _ ___ _| |_ ___  ___
 | |    | | | | / __| __/ _ \ '__| | '_ \| '__/ _ \ '__/ _ \/ _` | | | | / __| | __/ _ \/ __|
 | |____| | |_| \__ \ ||  __/ |    | |_) | | |  __/ | |  __/ (_| | |_| | \__ \ | ||  __/\__ \
  \_____|_|\__,_|___/\__\___|_|    | .__/|_|  \___|_|  \___|\__, |\__,_|_|___/_|\__\___||___/
                                   | |                         | |                           
                                   |_|                         |_|                           
```

## Overview

The `cluster-prerequisites` folder contains kubernetes resources which need to be installed before cluster setup.

Currently, these resources include the following `yaml` files:

- `default-sa-rbac-role.yaml` - This file binds the **cluster-admin** role with the default **ServiceAccount** to provide increased permissions that you require to complete the Kyma installation.

- `limit-range.yaml` - This file defines the memory limit range applied for kyma system namespaces. 
In case of an OOM error, adjust memory requirements for your components. Information about OOM error is on Pods, ReplicaSets, StatefulSets. 

- `resource-quotas.yaml` - This file defines resource quotas for the `kyma-system`, `kyma-integration` and `istio-system` Kyma system Namespaces.
If you receive an error during the Pod creation that relates to exceeding resource quota limits, adjust values in the file.

- `resource-quotas-installer.yaml` - This file contains resource quotas for the `kyma-installer` Namespace.

