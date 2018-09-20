```
  _              _                        _                      _
 | |            | |                      | |                    | |
 | |__     ___  | |  _ __ ___    ______  | |__    _ __    ___   | | __   ___   _ __   ______   _ __    ___   _ __     ___
 | '_ \   / _ \ | | | '_ ` _ \  |______| | '_ \  | '__|  / _ \  | |/ /  / _ \ | '__| |______| | '__|  / _ \ | '_ \   / _ \
 | | | | |  __/ | | | | | | | |          | |_) | | |    | (_) | |   <  |  __/ | |             | |    |  __/ | |_) | | (_) |
 |_| |_|  \___| |_| |_| |_| |_|          |_.__/  |_|     \___/  |_|\_\  \___| |_|             |_|     \___| | .__/   \___/
                                                                                                            | |
                                                                                                            |_|
```

## Overview

This directory contains the resources of the Kyma Helm Broker, which include scripts used in the process of developing yBundles and provisioning them in the Kyma cluster. The directory also contains the yBundles that you can provision in Kyma using the Helm Broker.

## Details
This section explains the structure of this directory and the purpose of its contents. It also shows the consequences of removing yBundles from the repository and the consequences of removing plans from yBundles.

### Resources
The Helm Broker-related resources are organized in these directories:

- [bundles](./bundles) which contains the yBundles definitions.
- [provisioning](./provisioning) which contains files which populate the Helm Broker with the yBundle definitions.
  - The [provision-bundles.sh](./provisioning/provision-bundles.sh) script creates the archives from the directories defined in the `bundles` directory. It also generates the `index.yaml` file and stores it together with the archives in a Persistent Volume used by the embedded bundles server. At the beginning, the Helm Broker reads the yBundle definitions from the embedded bundles server.
- [development](./development) which contains the scripts used during the yBundles development.
  - The [sync-bundles.sh](./development/sync-bundles.sh) script populates the bundles from the [bundles](./bundles) directory to Helm Broker embedded HTTP server repository. The script executes the `provision-bundles.sh` script and restarts the Helm Broker.
You can add and modify yBundles on your local machine using a preferred editor, and then execute the script to reflect local changes on the server.
You must wait one minute to synchronize these changes to the Service Catalog.
  - The [check.sh](./development/check.sh) script validates yBundles. It reads the yBundle definition and prints the basic information. It also executes the `helm install` command with a `dry-run` option set for every plan.
Run this command to test the included Redis yBundle using Tiller on Minikube. Set the **--kube-context** argument to `minikube`.
```
./development/check.sh --kube-context minikube --dry-run bundles/redis-0.0.3
```

  This is the list of the flags you can use with the `check.sh` script:

  ```
  -h --help           helm for the script
  -c --kube-context   kube context to use (required if the --dry-run is selected)
  --debug             prints the output from the Helm command
  --dry-run           performs the helm install --dry-run operation
  ```  

### Removing yBundles from the repository

If you remove a yBundle from which no ServiceBinding or ServiceClass were created, the ClusterServiceClass associated with that yBundle is removed from the Service Catalog.

If you remove a yBundle from which a ServiceBinding or a ServiceClass were created, the ClusterServiceClass associated with that yBundle is only marked as removed with the **status.removedFromBrokerCatalog** parameter set to `true`. In this case, the consequences are as follows:
- When you try to create a new ServiceInstance for the ClusterServiceClass, the Service Catalog does not trigger the Service Broker and instead marks the ServiceInstance as failed with `ReferencesDeletedServicePlan` error.
- You can create new ServiceBindings for the ServiceInstance associated with the yBundle you removed. You can also delete that ServiceInstance and all of its ServiceBindings.

When you remove all ServiceBindings and ServiceInstances associated with the deleted yBundle, the ClusterServiceClass is removed from the Service Catalog.


### Removing plans from yBundles

If you remove a plan for which no ServiceInstance was created, the ClusterServicePlan is removed from the ClusterServiceClass.

If you remove a plan for which a ServiceInstance was created, the associated ClusterServicePlan is is only marked as removed with the **status.removedFromBrokerCatalog** parameter set to `true`. In this case, the consequences are as follows:
- When you try to create a new ServiceInstance for the ClusterServicePlan, the Service Catalog does not trigger the Service Broker and instead marks the ServiceInstance as failed with `ReferencesDeletedServicePlan` error.
- You can create new ServiceBindings for the ServiceInstance associated with the yBundle you removed. You can also delete that ServiceInstance and all of its ServiceBindings.

When you remove all ServiceBindings and ServiceInstances associated with the deleted yBundle plan, the ClusterServicePlan is removed from the ClusterServiceClass.
