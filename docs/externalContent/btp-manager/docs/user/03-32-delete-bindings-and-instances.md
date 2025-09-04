# Delete Service Bindings and Service Instances

Delete service bindings and service instances using Kyma dashboard or kubectl.

## Prerequisites

* You have the [SAP BTP Operator module](README.md) added. For instructions on adding modules, see [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module).
* For CLI interactions: [kubectl](https://kubernetes.io/docs/tasks/tools/) configured to communicate with your Kyma instance. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?version=Cloud).

## Context

You can only delete service instances or service bindings created in Kyma using Kyma dashboard or kubectl. You can't perform these operations using the SAP BTP cockpit. To delete a service instance, first delete its service bindings.

> [!WARNING]
> Once you delete your service instances and service bindings, you cannot revert the operation.

## Procedure

Use either Kyma dashboard or kubectl to delete a service binding or a service instance.

Kyma dashboard is a web-based UI providing a graphical overview of your cluster and all its resources.
To access Kyma dashboard, use the link available in the **Kyma Environment** section of your subaccount **Overview**.

<!-- tabs:start -->

#### Kyma Dashboard

1. In the navigation area, choose **Namespaces**, and go to the namespace you want to delete a service binding/instance from.
2. Go to **Service Management**, and then to **Service Bindings** or **Service Instances**.
3. Choose the service binding/instance and delete it.

#### kubectl

To delete a service binding, run:

```bash
kubectl delete servicebindings.services.cloud.sap.com {BINDING_NAME}
```

To delete a service instance, run:

```bash
kubectl delete serviceinstances.services.cloud.sap.com {SERVICE_INSTANCE_NAME}
```

<!-- tabs:end -->
