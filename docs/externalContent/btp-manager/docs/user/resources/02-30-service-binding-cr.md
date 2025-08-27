# Service Binding Custom Resource

The `servicebindings.services.cloud.sap.com` CustomResourceDefinition (CRD) is a comprehensive specification that defines the structure and format used to configure a ServiceBinding resource.

To get the latest CRD in the YAML format, run the following command:

```shell
kubectl get crd servicebindings.services.cloud.sap.com -o yaml
```

## Example Custom Resource

The following ServiceBinding object is an example of a service binding configuration:

```yaml
apiVersion: services.cloud.sap.com/v1
kind: ServiceBinding
metadata:
  name: {BINDING_NAME}
spec:
  serviceInstanceName: {SERVICE_INSTANCE_NAME}
  externalName: {BINDING_NAME}
  secretName: {BINDING_NAME}
  parameters:
    key1: val1
    key2: val2      
```

## Custom Resource Parameters

The following table lists the parameters of the given resource with their descriptions:

**Spec:**

| Parameter             | Type   | Description                                                                                                                                    |
|-------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------|
| **serviceInstanceName** | string | The Kubernetes name of the service instance to bind. |
| **serviceInstanceNamespace** | string | The namespace of the service instance to bind; if not specified, the default is the binding's namespace. |
| **externalName**       | string  | The name for the service binding in SAP BTP; if not specified, defaults to the binding **metadata.name**. |
| **secretName**         | string  | The name of the Secret where the credentials are stored; if not specified, defaults to the binding **metadata.name**. |
| **secretKey**          | string  | The Secret key is part of the Secret object, which stores the service binding credentials received from the service broker. When the Secret key is used, all the credentials are stored under a single key. This makes it a convenient way to store credentials data in one file when using volumeMounts. See [Formatting Service Binding Secret](../03-50-formatting-service-binding-secret.md). |
| **secretRootKey**       | string | The root key is part of the Secret object, which stores the service binding credentials received from the service broker, and additional service instance information. When the root key is used, all data is stored under a single key. This makes it a convenient way to store data in one file when using volumeMounts. See [Formatting Service Binding Secret](../03-50-formatting-service-binding-secret.md). |
| **parameters**          | []object | Some services support the provisioning of additional configuration parameters during the bind request.<br/>For the list of supported parameters, check the documentation of particular service offerings. |
| **parametersFrom**      | []object | List of sources from which parameters are populated. |
| **userInfo**            | object | Contains information about the user that last modified this service binding. |
| **credentialsRotationPolicy** | object | Holds automatic credentials rotation configuration.  |
| **credentialsRotationPolicy.enabled** | boolean  | Indicates whether automatic credentials rotation is enabled. |
| **credentialsRotationPolicy.rotationFrequency** | duration | Specifies the frequency at which the binding rotation is performed. |
| **credentialsRotationPolicy.rotatedBindingTTL** | duration | Specifies the period for which the rotated binding is kept. |
| **SecretTemplate**      | string | A Go template used to generate a custom Kubernetes v1/Secret, working on both the access credentials returned by the service broker and instance attributes. See [Go Templates](https://pkg.go.dev/text/template) for more details. |

**Status:**

| Parameter         | Type     | Description                                                                                                   |
|-----------------|---------|-----------------------------------------------------------------------------------------------------------|
| **instanceID**   | string | The ID of the bound instance in the SAP Service Manager service. |
| **bindingID**    | string | The service binding ID in the SAP Service Manager service. |
| **operationURL** | string | The URL of the current operation performed on the service binding. |
| **operationType**| string | The type of the current operation. Possible values are `CREATE`, `UPDATE`, or `DELETE`. |
| **conditions** | []condition | An array of conditions describing the status of the service instance.<br>The possible conditions types are:<ul><li>`Ready:true` if the binding is ready and usable</li><li>`Failed:true` when an operation on the service binding fails. In the case of failure, the details about the error are available in the condition message.</li><li>`Succeeded:true` when an operation on the service binding succeeded. If set to `false`, the operation is considered in progress unless a `Failed` condition exists.</li></ul> |
| **lastCredentialsRotationTime**| time | Indicates the last time the binding secret was rotated. |
