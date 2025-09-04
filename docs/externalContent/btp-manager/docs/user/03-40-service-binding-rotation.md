# Rotate Service Bindings

Enhance security by automatically rotating the credentials associated with your service bindings. This process involves generating a new service binding while keeping the old credentials active for a specified period to ensure a smooth transition.

## Enable Automatic Rotation

To enable automatic service binding rotation, use the **credentialsRotationPolicy** field within the `spec` section of the ServiceBinding resource. You can configure the following parameters:

| Parameter         | Type     | Description                                                                                                                               | Valid Values |
|-----------------------|---------|----------------------------------------------------------------------------------------------------------------------------------------|--------------|
| **enabled**           | bool    | Turns automatic rotation on or off.                                                                                                    | `true` or `false`                    |
| **rotationFrequency** | string  | Defines the desired interval between binding rotations.             | "m" (minutes), "h" (hours)|
| **rotatedBindingTTL** | string  | Determines how long to keep the old ServiceBinding resource after rotation and before deletion. The actual TTL may be slightly longer. | "m" (minutes), "h" (hours) |   

> [!NOTE] 
> The `credentialsRotationPolicy` does not manage the validity or expiration of the credentials themselves. This is determined by the service you are using.

The `credentialsRotationPolicy` is evaluated periodically during a [control loop](https://kubernetes.io/docs/concepts/architecture/controller/) on every service binding update or during a complete reconciliation process. This means the actual rotation occurs in the closest upcoming reconciliation loop. 

## Enable Immediate Rotation

To trigger an immediate rotation regardless of the configured rotation frequency, add the `services.cloud.sap.com/forceRotate: "true"` annotation to the ServiceBinding resource.
The immediate rotation only works if automatic rotation is already enabled. 

The following example shows the configuration of a ServiceBinding resource for rotating credentials every 25 days (600 hours) and keeping the old ServiceBinding resource for 2 days (48 hours) before deleting it:

```yaml
apiVersion: services.cloud.sap.com/v1
kind: ServiceBinding
metadata:
  name: {BINDING_NAME}
spec:
  serviceInstanceName: {SERVICE_INSTANCE_NAME}
  credentialsRotationPolicy:
    enabled: true
    rotatedBindingTTL: 48h
    rotationFrequency: 600h
 ```

## Result

Rotating the service binding has the following results:

* The Secret is updated with the latest credentials. 
* The old credentials are kept in a newly-created Secret named `original-secret-name(variable)-guid(variable)`.
This temporary Secret is kept until the configured deletion time (TTL).

To see the timestamp of the last service binding rotation, go to the **status.lastCredentialsRotationTime** field.

## Limitations 

You cannot enable automatic credential rotation for a backup service binding (named: original-binding-name(variable)-guid(variable)) marked with the `services.cloud.sap.com/stale` label.
This backup service binding is created during the credentials rotation process to facilitate the process.
