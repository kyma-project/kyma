# Namespace-Level Mapping

You can map a Kubernetes namespace to an SAP Service Manager instance in a given subaccount. The Service Manager instance is then used to provision all service instances in that namespace.

## Prerequisites

* A subaccount in the SAP BTP cockpit.
* You have the [SAP BTP Operator module](README.md) added. For instructions on adding modules, see [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module).
* [kubectl](https://kubernetes.io/docs/tasks/tools/) configured to communicate with your Kyma instance. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?version=Cloud).

## Context

To connect a namespace to a specific subaccount, maintain the access credentials to the subaccount in a Secret dedicated to a specific namespace. Create the `{NAMESPACE-NAME}-sap-btp-service-operator` Secret in the `kyma-system` namespace.

## Create a Namespace-Based Secret

1. In the SAP BTP cockpit, create a new SAP Service Manager service instance with the `service-operator-access` plan. See [Creating Instances in Other Environments](https://help.sap.com/docs/service-manager/sap-service-manager/creating-instances-in-other-environments?locale=en-US&version=Cloud).
2. Create a service binding to the SAP Service Manager service instance you have created. See [Creating Service Bindings in Other Environments](https://help.sap.com/docs/service-manager/sap-service-manager/creating-service-bindings-in-other-environments?locale=en-US&version=Cloud).
3. Get the access credentials of the SAP Service Manager instance from its service binding. Copy them from the SAP BTP cockpit as a JSON file.
4. Create the `creds.json` file in your working directory and save the credentials there.
5. In the same working directory, call the `create-secret-file.sh` script with the **operator** option as the first parameter and **namespace-name-sap-btp-service-operator** Secret as the second parameter.

    ```sh
    curl https://raw.githubusercontent.com/kyma-project/btp-manager/main/hack/create-secret-file.sh | bash -s operator {NAMESPACE_NAME}-sap-btp-service-operator
    ```

    The expected result is the `btp-access-credentials-secret.yaml` file created in your working directory:

    ```yaml
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: {NAMESPACE_NAME}-sap-btp-service-operator
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

   You see the status `Created`.


## Create a Service Instance with a Namespace-Based Secret

1. To create a service instance with a namespace-based Secret, follow the instructions in [Create Service Instances and Service Bindings](03-30-create-instances-and-bindings.md).

2. To verify that you've correctly added the access credentials of the SAP Service Manager instance in your service instance, go to the custom resource (CR) `status` section, and make sure the subaccount ID to which the instance belongs is provided in the **subaccountID** field. The field must not be empty.

## Related Information

[Working with Multiple Subaccounts](03-20-multitenancy.md)<br>
[Instance-Level](03-21-instance-level-mapping.md)
