# Create Service Instances and Service Bindings

To use an SAP BTP service in your Kyma cluster, create its service instance and service binding using Kyma dashboard or kubectl.

## Prerequisites

* You have the [SAP BTP Operator module](README.md) added. For instructions on adding modules, see [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module).
* For CLI interactions: [kubectl](https://kubernetes.io/docs/tasks/tools/) configured to communicate with your Kyma instance. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?version=Cloud).
* For an enterprise account, you have added quotas to the services you purchased in your subaccount. Otherwise, only default free-of-charge services are listed in the service marketplace. Quotas are automatically assigned to the resources available in trial accounts.
  For more information, see [Configure Entitlements and Quotas for Subaccounts](https://help.sap.com/docs/btp/sap-business-technology-platform/configure-entitlements-and-quotas-for-subaccounts?&version=Cloud).
* You know the service offering name and service plan name for the SAP BTP service you want to connect to your Kyma cluster.
  >[!TIP]
  >To find the service and service plan names, in the SAP BTP cockpit, go to **Services**->**Service Marketplace**. Click on the service tile and find its **name** and **Plan**.

## Create a Service Instance

To create a service instance, use either Kyma dashboard or kubectl.

<!-- tabs:start -->

#### Kyma Dashboard

Kyma dashboard is a web-based UI providing a graphical overview of your cluster and all its resources.
To access Kyma dashboard, use the link available in the **Kyma Environment** section of your subaccount **Overview**.

1. In the navigation area, choose **Namespaces**, and go to the namespace you want to work in.
2. Go to **Service Management** -> **Service Instances**, and choose **Create**.
3. Provide the required service details in **Form**. Alternatively, you can switch to the **YAML** tab and edit or upload your file.
4. Choose **Create**.

   You see the status `PROVISIONED`.
   
#### kubectl

1.  To create a ServiceInstance custom resource (CR), replace the placeholders and run the following command:

    ```yaml
    kubectl create -f - <<EOF 
    apiVersion: services.cloud.sap.com/v1
    kind: ServiceInstance
    metadata:
        name: {SERVICE_INSTANCE_NAME}
        namespace: {NAMESPACE} 
    spec:
        serviceOfferingName: {SERVICE_OFFERING_NAME}
        servicePlanName: {SERVICE_PLAN_NAME}
        externalName: {SERVICE_INSTANCE_NAME}
        parameters:
          key1: val1
          key2: val2
    EOF
    ```
    
2.  To check the service's status in your cluster, run:
   
    ```bash
    kubectl get serviceinstances.services.cloud.sap.com {SERVICE_INSTANCE_NAME} -n {NAMESPACE}
    ```

    You get an output similar to this one:

    ```
    NAME                      OFFERING                  PLAN                  STATUS    AGE
    {SERVICE_INSTANCE_NAME}   {SERVICE_OFFERING_NAME}   {SERVICE_PLAN_NAME}   Created   44s
    ```

<!-- tabs:end -->

## Create a Service Binding

To create a service binding, use either Kyma dashboard or kubectl.

With a ServiceBinding custom resource (CR), your application can get access credentials for communicating with an SAP BTP service.
These access credentials are available to applications through a Secret resource generated in your cluster.

### Procedure

<!-- tabs:start -->

#### Kyma Dashboard

Kyma dashboard is a web-based UI providing a graphical overview of your cluster and all its resources.
To access Kyma dashboard, use the link available in the **Kyma Environment** section of your subaccount **Overview**.

1. In the navigation area, choose **Namespaces**, and go to the namespace you want to work in.
2. Go to **Service Management** -> **Service Bindings**, and choose **Create**.
3. Provide the required details, and choose your service instance name from the dropdown list. Alternatively, you can provide the required details by switching from the **Form** to **YAML** tab, and editing or uploading your file.

4. Choose **Create**.

   You see the status `PROVISIONED`.

#### kubectl

1. To create a ServiceBinding CR, replace the palceholders and run the following command:

      ```yaml
      kubectl create -f - <<EOF
      apiVersion: services.cloud.sap.com/v1
      kind: ServiceBinding
      metadata:
        name: {BINDING_NAME}
      spec:
        serviceInstanceName: {SERVICE_INSTANCE_NAME}
        externalName: {EXTERNAL_NAME}
        secretName: {SECRET_NAME}
        parameters:
          key1: val1
          key2: val2   
      EOF        
      ```

    > [!NOTE]
    > In the **serviceInstanceName** field of the ServiceBinding, enter the name of the ServiceInstance resource you previously created.
    
2.  To check your service binding status, run:

    ```bash
    kubectl get servicebindings {BINDING_NAME} -n {NAMESPACE}
    ```

    You see the staus `Created`.

3.  Verify the Secret is created with the name specified in the  **spec.secretName** field of the ServiceBinding CR. The Secret contains access credentials that the applications need to use the service:

    ```bash
    kubectl get secrets {SECRET_NAME} -n {NAMESPACE}
    ```
    You see the same Secret name as in the **spec.secretName** field of the ServiceBinding CR.

<!-- tabs:end -->

### Results

You can use a given service in your Kyma cluster.
