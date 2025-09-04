# Migrate Service Instances and Service Bindings from a Custom SAP BTP Service Operator
Learn how to migrate service instances and service bindings from a custom SAP BTP service operator to a Kyma cluster.

## Prerequsites

* You have a Kyma cluster deployed
* You have the SAP BTP Operator module added. For instructions on adding modules, see [Adding and Deleting a Kyma Module](https://help.sap.com/docs/btp/sap-business-technology-platform/enable-and-disable-kyma-module?version=Cloud).
* For CLI interactions: [kubectl](https://kubernetes.io/docs/tasks/tools/) configured to communicate with your Kyma instance. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?version=Cloud).
* For processing JSON data output from kubectl commands: [jq](https://jqlang.org/)

## Context

If you have service instances and service bindings created outside the Kyma environment, and you want to continue using them in the Kyma environment, you must migrate them into your Kyma cluster.

## Prepare Migration Data from Your Custom SAP BTP Service Operator

1. To connect kubectl to the cluster with your custom SAP BTP service operator, set either the **KUBECONFIG** environment variable or the cluster context with the `kubectl config use-context` command.
2. Find the `sap-btp-service-operator` Secret by running the `kubectl get secrets -A` command.
3. Save the Secret name and its namespace in the **SAP_BTP_OPERATOR_SECRET_NAME** and **SAP_BTP_OPERATOR_SECRET_NAMESPACE** environment variables.
4. Save the SAP BTP service operator credentials in the following environment variables:

    ```
    CLIENT_ID=$(kubectl get secret -n $SAP_BTP_OPERATOR_SECRET_NAMESPACE $SAP_BTP_OPERATOR_SECRET_NAME -o jsonpath={.data.clientid})
    CLIENT_SECRET=$(kubectl get secret -n $SAP_BTP_OPERATOR_SECRET_NAMESPACE $SAP_BTP_OPERATOR_SECRET_NAME -o jsonpath={.data.clientsecret})
    SM_URL=$(kubectl get secret -n $SAP_BTP_OPERATOR_SECRET_NAMESPACE $SAP_BTP_OPERATOR_SECRET_NAME -o jsonpath={.data.sm_url})
    TOKEN_URL=$(kubectl get secret -n $SAP_BTP_OPERATOR_SECRET_NAMESPACE $SAP_BTP_OPERATOR_SECRET_NAME -o jsonpath={.data.tokenurl})
    ```

5. To find the `sap-btp-operator-config` ConfigMap in the cluster, run the `kubectl get configmaps -A` command.
6. Save the ConfigMap name and its namespace in the **SAP_BTP_OPERATOR_CONFIGMAP_NAME** and **SAP_BTP_OPERATOR_CONFIGMAP_NAMESPACE** environment variables.
7. Save the cluster ID in the environment variable:

    ```
    CLUSTER_ID=$(kubectl get configmap -n $SAP_BTP_OPERATOR_CONFIGMAP_NAMESPACE $SAP_BTP_OPERATOR_CONFIGMAP_NAME -o jsonpath={.data.CLUSTER_ID} | base64)
    ```

8. List all service instances with the `kubectl get serviceinstances -A` command. Take note of the namespaces that must be present in the Kyma cluster.
   When you migrate service instances into the Kyma cluster, you must use these noted namespaces to recreate the necessary structure within the new environment.
9. Save each service instance you want to migrate as a manifest in a JSON file. To do that, replace the placeholders with the actual service instance name and its namespace, and run:

    ```
    kubectl get serviceinstance -n <SERVICE_INSTANCE_NAMESPACE> <SERVICE_INSTANCE_NAME> -o json \
    | jq 'del(.metadata.annotations, .metadata.creationTimestamp, .metadata.finalizers, .metadata.generation, .metadata.resourceVersion, .metadata.uid, .status)' \
    > <SERVICE_INSTANCE_NAME>-si.json
    ```

10. List all service bindings with the `kubectl get servicebindings -A` command. Take note of namespaces that must be present in the Kyma cluster.
    When you migrate service bindings into the Kyma cluster, you must use these noted namespaces to recreate the necessary structure within the new environment. 
11. Save each service binding you want to migrate as a manifest in a JSON file. To do that, replace the placeholders with the actual service binding name and its namespace, and run:

    ```
    kubectl get servicebinding -n <SERVICE_BINDING_NAMESPACE> <SERVICE_BINDING_NAME> -o json \
    | jq 'del(.metadata.annotations, .metadata.creationTimestamp, .metadata.finalizers, .metadata.generation, .metadata.ownerReferences, .metadata.resourceVersion, .metadata.uid, .status)' \
    > <SERVICE_BINDING_NAME>-sb.json
    ```

## Migrate Resources to a Kyma Cluster

To migrate your resources to a Kyma cluster, you must first customize the `sap-btp-manager` Secret. To prevent automatic reversion of your custom changes, add the `kyma-project.io/skip-reconciliation: 'true'` label to the Secret and perform the following steps:

1. To connect kubectl to your Kyma cluster, set either the **KUBECONFIG** environment variable or the cluster context with the `kubectl config use-context` command.
2. To find the `sap-btp-manager` Secret in the Kyma cluster, run the `kubectl get secrets -A` command.
3. Save the Secret name and its namespace in the **BTP_MANAGER_SECRET_NAME** and **BTP_MANAGER_SECRET_NAMESPACE** environment variables.
4. When preparing resources for migration, you noted the namespaces that must be present in the Kyma cluster. To avoid errors and maintain structures based on specific namespace assignments, recreate these namespaces required for the service instances and service bindings you are migrating.
5. To override the SAP BTP service operator credentials and cluster ID, patch the `sap-btp-manager` Secret:

    ```
    kubectl patch secret -n ${BTP_MANAGER_SECRET_NAMESPACE} ${BTP_MANAGER_SECRET_NAME} -p "{\"data\":{\"clientid\":\"${CLIENT_ID}\",\"clientsecret\":\"${CLIENT_SECRET}\",\"sm_url\":\"${SM_URL}\",\"tokenurl\":\"${TOKEN_URL}\",\"cluster_id\":\"${CLUSTER_ID}\"}}"
    ```

6. To create service instances and service bindings from the saved manifests in JSON format, use the `kubectl apply -f <JSON_MANIFEST>` command. **JSON_MANIFEST** is a placeholder for the actual service instance or service binding JSON manifest.

## Deleting Migrated Resources

To limit service instances and service bindings management to the SAP BTP service operator in the Kyma cluster, perform the following steps:

1. To avoid deleting the resources from your Kyma cluster, ensure that you connect kubectl to the cluster with your custom SAP BTP service operator. Set either the **KUBECONFIG** environment variable or the cluster context with the `kubectl config use-context` command.
2. To find the `sap-btp-operator-controller-manager` deployment in the cluster with your custom SAP BTP service operator, run the `kubectl get deployments -A` command.
3. Save the deployment name and its namespace in the **SAP_BTP_OPERATOR_DEPLOYMENT_NAME** and **SAP_BTP_OPERATOR_DEPLOYMENT_NAMESPACE** environment variables.
4. To scale the deployment to 0 replicas, run: 

    ```
    kubectl scale deployment -n $SAP_BTP_OPERATOR_DEPLOYMENT_NAMESPACE $SAP_BTP_OPERATOR_DEPLOYMENT_NAME --replicas=0
    ```

5. To delete the SAP BTP service operator webhooks, run:

    ```
    kubectl delete mutatingwebhookconfigurations sap-btp-operator-mutating-webhook-configuration && kubectl delete validatingwebhookconfigurations sap-btp-operator-validating-webhook-configuration
    ```

6. Delete finalizers from each migrated service instance and service binding.
7. Delete migrated service bindings.
8. Delete migrated service instances.

## Results

Your service instances and service bindings are now migrated from another environment to the Kyma environment, and you manage them from the SAP BTP service operator in your Kyma cluster.

## Read More
[Customize the Default Credentials and Access](03-11-customize_secret.md)