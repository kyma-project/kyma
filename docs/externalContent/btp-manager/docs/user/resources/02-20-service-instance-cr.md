# Service Instance Custom Resource

The `serviceinstances.services.cloud.sap.com` CustomResourceDefinition (CRD) is a comprehensive specification that defines the structure and format used to configure a ServiceInstance resource.

To get the latest CRD in the YAML format, run the following command:

```shell
kubectl get crd serviceinstances.services.cloud.sap.com -o yaml
```

## Example Custom Resource

The following ServiceInstance object is an example of a service instance configuration:

```yaml
    apiVersion: services.cloud.sap.com/v1
    kind: ServiceInstance
    metadata:
        name: {SERVICE_INSTANCE_NAME}
    spec:
        serviceOfferingName: {SERVICE_OFFERING_NAME}
        servicePlanName: {SERVICE_PLAN_NAME}
        externalName: {SERVICE_INSTANCE_NAME}
        parameters:
          key1: val1
          key2: val2
```

## Custom Resource Parameters

The following table lists the parameters of the given resource with their descriptions:

**Spec:**

| Parameter             | Type   | Description                                                                                                                                    |
|-------------------------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| **serviceOfferingName** | string    | The name of the SAP BTP service you want to consume. |
| **servicePlanName**     | string    | The plan of the selected service offering you want to consume. |
| **servicePlanID**        | string   | The plan ID. |
| **externalName**         | string   | The name for the service instance in SAP BTP; if not specified, defaults to the instance **metadata.name**. |
| **parameters**           | []object | Some services support the provisioning of additional configuration parameters during the instance creation.<br/>For the list of supported parameters, check the documentation of the particular service offering. |
| **parametersFrom**       | []object | List of sources from which parameters are populated. |
| **watchParametersFromChanges** | bool | If set to `true`, it triggers an automatic update of the ServiceInstance resource with the changes to the Secret values listed in **parametersFrom**. Use this field together with **parameterFrom**.<br>Defaults to `false`. |
| **customTags**           | []string | List of custom tags describing the service instance; copied to the ServiceBinding Secret in the key called **tags**. |
| **userInfo**             | object   | The user that last modified this service instance. |
| **shared**               | *bool    | The shared state. Possible values: `true`, `false`, or none.<br> If not specified, the value defaults to `false`. |
| **btpAccessCredentialsSecret** | string   | The name of the Secret that contains access credentials for the SAP BTP service operator. See [Working with Multiple Subaccounts](../03-20-multitenancy.md). |

**Status:**

| Parameter         | Type     | Description                                                                                                   |
|-----------------|---------|-----------------------------------------------------------------------------------------------------------|
| **instanceID**   | string | The service instance ID in the SAP Service Manager service.  |
| **operationURL** | string | The URL of the current operation performed on the service instance.  |
| **operationType** | string | The type of the current operation. Possible values are `CREATE`, `UPDATE`, or `DELETE`. |
| **conditions**   | []condition | An array of conditions describing the status of the service instance.<br/>The possible condition types are:<ul><li>`Ready: true` if the instance is ready and usable</li><li>`Failed: true` when an operation on the service instance fails.</li><li>`Succeeded: true` when an operation on the service instance succeeded. If set to `false`, it is considered in progress unless a `Failed` condition exists.</li><li>`Shared: true` when sharing of the service instance succeeded. If set to `false`, unsharing of the service instance succeeded or the service instance is not shared.</li></ul> |
| **tags**       | []string   | Tags describing the service instance as provided in the service catalog, which is copied to the service binding Secret in the key called **tags**.|

## Annotations

| Parameter         | Type                 | Description                                                                                                                                                                                                     |
|-----------------|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| services.cloud.sap.com/preventDeletion   | map[string] string | You can prevent deletion of any service instance by adding the following annotation: `services.cloud.sap.com/preventDeletion : "true"`.<br>To enable back the deletion of the instance, either remove the annotation or set it to `false`. |