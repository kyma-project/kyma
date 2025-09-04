# Create an SAP BTP Service Instance in Your Kyma Cluster

After successfully installing your Secret, create a service instance and a service binding.

> [!NOTE] 
> This section provides an example with the Authorization and Trust Management (`xsuaa`) service. Create your Secret following this example:

## Procedure

1. To create a service instance, run the following script:

    ```yaml
    kubectl create -f - <<EOF
    apiVersion: services.cloud.sap.com/v1
    kind: ServiceInstance
    metadata:
      name: {SERVICE_INSTANCE_NAME}
      namespace: default
    spec:
      serviceOfferingName: xsuaa
      servicePlanName: application
      externalName: {SERVICE_INSTANCE_NAME}
    EOF
    ```

   > [!TIP] 
   > To find values for the **serviceOfferingName** and **servicePlanName** parameters, go to the SAP BTP cockpit > **Service Marketplace**, select the service's tile, and find the **name** and **Plan**. The value of the **externalName** parameter must be unique.

2. To check the output, run:

    ```bash
    kubectl get serviceinstances.services.cloud.sap.com {SERVICE_INSTANCE_NAME} -o yaml
    ```

    You see the status `Created` and the message `ServiceInstance provisioned successfully`.

3. To create a service binding, run this script:

    ```yaml
    kubectl create -f - <<EOF
    apiVersion: services.cloud.sap.com/v1
    kind: ServiceBinding
    metadata:
      name: {BINDING_NAME}
      namespace: default
    spec:
      serviceInstanceName: {SERVICE_INSTANCE_NAME}
      externalName: {BINDING_NAME}
      secretName: {BINDING_NAME}
    EOF
    ```

4. To check the output, run:

    ```bash
    kubectl get servicebindings.services.cloud.sap.com {BINDING_NAME} -o yaml
    ```

    You see the status `Created` and the message `ServiceBinding provisioned successfully`.

5. Now, use a given service in your Kyma cluster. To see credentials, run:

    ```bash
    kubectl get secret {BINDING_NAME} -o yaml
    ```

## Result

You can use the Secret to communicate with the service instance.
