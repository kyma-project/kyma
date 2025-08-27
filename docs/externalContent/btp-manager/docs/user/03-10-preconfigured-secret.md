# Preconfigured Credentials and Access

When you create SAP BTP, Kyma runtime, all necessary resources for consuming SAP BTP services are created, and the basic cluster access is configured.

## Credentials

When you create a Kyma instance in the SAP BTP cockpit, the following events happen in your subaccount:

1. An SAP Service Manager service instance with the `service-operator-access` plan is created.
2. An SAP Service Manager service binding with access credentials for the SAP BTP Operator is created.
3. The credentials from the service binding are passed on to the Kyma service instance in the creation process.
4. The `sap-btp-manager` Secret is created and managed in the `kyma-system` namespace.
5. By default, the SAP BTP Operator module is installed together with the following:

   * The `sap-btp-manager` Secret.
   * The `sap-btp-service-operator` Secret with the access credentials for the SAP BTP service operator. You can view the credentials in the `kyma-system` namespace.
   * The `sap-btp-operator-config` ConfigMap.

> [!TIP] <!--OS only-->
> In this scenario, the `sap-btp-service-operator` Secret is automatically generated when you create Kyma runtime. To create this Secret manually for a specific namespace, see [Create a Namespace-Based Secret](03-22-namespace-level-mapping.md#create-a-namespace-based-secret).

The `sap-btp-manager` Secret provides the following credentials:

* **clientid**
* **clientsecret**
* **cluster_id**
* **sm_url**
* **tokenurl**

> [!NOTE]
> If you modify or delete the `sap-btp-manager` Secret, it is modified back to its previous settings or regenerated within 24 hours.
> To prevent your changes from being reverted, label the Secret with `kyma-project.io/skip-reconciliation: "true"`. For more information, see [Customize the Default Credentials and Access](03-11-customize_secret.md).

When you add the SAP BTP Operator module to your cluster, the `sap-btp-manager` Secret populates the SAP BTP service operator's resources as shown in the following diagram:
<!-- for the HP doc this sentence is different: The SAP BTP Operator module is added by default to your cluster and the `sap-btp-manager` (...) -->

![module_credentials](../assets/module_credentials.drawio.svg)

The cluster ID represents and identifies a Kyma service instance created in a particular subaccount. In Kyma dashboard, you can view the cluster ID in the following resources:

* The `sap-btp-manager` Secret
* The `sap-btp-service-operator` Secret
* The `sap-btp-operator-config` ConfigMap

## Cluster Access

By default, SAP BTP Operator has cluster-wide permissions.

The following parameters manage cluster access:

| Parameter                | Description                                                                                                                                                                              |
|--------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **CLUSTER_ID**           | Generated when Kyma runtime is created.                                                                                                                                                  |
| **MANAGEMENT_NAMESPACE** | Indicates the namespace for the Secrets with credentials to communicate with the SAP Service Manager. By default, set to `kyma-system`. |
| **RELEASE_NAMESPACE**    | Indicates the namespace for the `sap-btp-service-operator` and `sap-btp-operator-clusterid` Secrets. By default, set to `kyma-system`.              |
| **ALLOW_CLUSTER_ACCESS** | You can use every namespace for your operations. The parameter is always set to `true`. If you change it to `false`, your setting is automatically reverted, and SAP BTP Operator cluster-wide permissions remain unchanged.                              |

## Default Credentials and Kyma Runtime Deletion

The preconfigured credentials described in the [Credentials](#credentials) section may affect the deletion of your Kyma cluster.

If any non-deleted service instances in your Kyma cluster use the credentials from the SAP Service Manager resources created automatically, you can't delete the cluster. In this case, the existing service instances block the cluster's deletion. Before you attempt to delete the cluster from the SAP BTP cockpit, delete your service instances and bindings in Kyma dashboard. See [Delete Service Bindings and Service Instances](03-32-delete-bindings-and-instances.md#kyma-dashboard).

However, if all existing service instances in your Kyma cluster use your custom SAP Service Manager credentials, the non-deleted service instances do not block the cluster's deletion. See [Customize the Default Credentials and Access](03-11-customize_secret.md#procedure).

If you have not deleted service instances and bindings connected to your expired free tier service, you can still find the service binding credentials in the SAP Service Manager instance details in the SAP BTP cockpit. Use them to delete the leftover service instances and bindings.


## Related Information

[Customize Default Credentials and Access](03-11-customize_secret.md)
