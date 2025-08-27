# Instance-Level Mapping

You can map a Kubernetes service instance to an SAP Service Manager instance in a given subaccount. The Service Manager instance is then used to provision that service instance.

## Prerequisites

* A subaccount in the SAP BTP cockpit.
* You have the [SAP BTP Operator module](README.md) added. For instructions on adding modules, see [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module).
* [kubectl](https://kubernetes.io/docs/tasks/tools/) configured to communicate with your Kyma instance. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?version=Cloud).

## Context

To have multiple service instances from different subaccounts associated with one namespace, you must store access credentials for each subaccount in a custom Secret in the `kyma-system` namespace.
To create a service instance with a custom Secret, you must use the **btpAccessCredentialsSecret** field in the `spec` of the service instance. In it, you pass the Secret from the `kyma-system` namespace to create your service instance. You can use different Secrets for different service instances.

## Create Your Custom Secret

1. In the SAP BTP cockpit, create an SAP Service Manager service instance with the `service-operator-access` plan. See [Creating Instances in Other Environments](https://help.sap.com/docs/service-manager/sap-service-manager/creating-instances-in-other-environments?locale=en-US&version=Cloud).
2. Create a service binding to the SAP Service Manager service instance you have created. See [Creating Service Bindings in Other Environments](https://help.sap.com/docs/service-manager/sap-service-manager/creating-service-bindings-in-other-environments?locale=en-US&version=Cloud).
3. Get the access credentials of the SAP Service Manager instance from its service binding. Copy them from the BTP cockpit as a JSON file.
4. Create the `creds.json` file in your working directory and save the credentials there.
5. In the same working directory, generate the Secret by calling the `create-secret-file.sh` script with the **operator** option as the first parameter and **your-secret-name** as the second parameter:

    ```sh
    curl https://raw.githubusercontent.com/kyma-project/btp-manager/main/hack/create-secret-file.sh | bash -s operator {YOUR_SECRET_NAME}
    ```

    The expected result is the file `btp-access-credentials-secret.yaml` created in your working directory:

    ```yaml
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: {YOUR_SECRET_NAME}
      namespace: kyma-system
    data:
      clientid: {CLIENT_ID}
      clientsecret: {CLIENT_SECRET}
      sm_url: {SM_URL}
      tokenurl: {AUTH_URL}
      tokenurlsuffix: "/oauth/token"
    ```

6. To create the Secret, run:

    ```
    kubectl create -f ./btp-access-credentials-secret.yaml
    ```

7. To verify that the Secret has been successfully created, run:
   
   ```
   kubectl get secret -n kyma-system {YOUR_SECRET_NAME}
   ```

   You see the status `Created`.

   > [!NOTE]
   > You can also view the Secret in Kyma dashboard. In the `kyma-system` namespace, go to **Configuration** -> **Secrets**, and check the list of Secrets.

## Create a Service Instance with the Custom Secret

To create the service instance, use either Kyma dashboard or kubectl.

Kyma dashboard is a web-based UI providing a graphical overview of your cluster and all its resources.
To access Kyma dashboard, use the link available in the **Kyma Environment** section of your subaccount **Overview**.

### Procedure

<Tabs>
<Tab name="Kyma Dashboard">

1. In the **Namespaces** view, go to the namespace you want to work in.
2. Go to **Service Management** -> **Service Instances**.
3. In the **BTP Access Credentials Secret** field, add the name of the custom Secret you have created.
4. Provide other required service details and create a service instance.

   > [!WARNING]
   > Once you set a Secret name in the service instance, you cannot change it in the future.
  
    You see the status `PROVISIONED`.
</Tab>
<Tab name="kubectl">

1. Create your service instance with the following:

   * The **btpAccessCredentialsSecret** field in the `spec` pointing to the custom Secret you have created
   * Other parameters as needed
    
    > [!WARNING] 
    > Once you set a Secret name in the service instance, you cannot change it in the future.

    See an example of a ServiceInstance custom resource (CR):

    ```yaml
    kubectl create -f - <<EOF
    apiVersion: services.cloud.sap.com/v1
    kind: ServiceInstance
    metadata:
      name: {SERVICE_INSTANCE_NAME}
      namespace: {NAMESPACE_NAME}
    spec:
      serviceOfferingName: {SERVICE_OFFERING_NAME}
      servicePlanName: {SERVICE_PLAN_NAME}
      btpAccessCredentialsSecret: {YOUR_SECRET_NAME}
    EOF
    ```

2. To verify that your service instance has been created successfully, run:

    ```bash
    kubectl get serviceinstances.services.cloud.sap.com {SERVICE_INSTANCE_NAME} -n {NAMESPACE}
    ```

    You see the status `Created` and the message that your service instance has been created successfully.
    You also see your Secret name in the **btpAccessCredentialsSecret** field of the `spec`.

3.  To verify that you've correctly added the access credentials of the SAP Service Manager instance in your service instance, go to the CR `status` section, and make sure the subaccount ID to which the instance belongs is provided in the **subaccountID** field. The field must not be empty.
</Tab>
</Tabs>
## Related Information

[Working with Multiple Subaccounts](03-20-multitenancy.md)<br>
[Namespace-Level Mapping](03-22-namespace-level-mapping.md)
