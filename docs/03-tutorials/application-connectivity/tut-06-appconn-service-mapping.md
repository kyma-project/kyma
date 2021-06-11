---
title: Bind a service to a Namespace
type: Tutorials
---

This guide shows how to bind a service provided by an external solution to a Namespace in Kyma. To execute the binding, provision a Service Instance for the registered service in the Namespace. Follow the instructions to bind your service to the desired Namespace.

> **NOTE:** To learn more about Service Instances and the Service Catalog domain in Kyma, see the [Service Catalog](/components/service-catalog/) topic.

## Prerequisites

- An Application created in your cluster and bound to the desired Namespace.
- A service provided by an external solution registered in the Application.

## Steps

1. Export the name of the desired Namespace and the name of the Service Instance.

   ```bash
   export NAMESPACE={DESIRED_NAMESPACE}
   export INSTANCE_NAME={SERVICE_INSTANCE_NAME}
   ```

2. Expose the `externalName` of the Service Class of the registered service.

   ```bash
   export EXTERNAL_NAME=$(kubectl -n $NAMESPACE get serviceclass {SERVICE_ID} -o jsonpath='{.spec.externalName}')
   ```

   > **NOTE:** `{SERVICE_ID}` is the identifier of the registered service.

3. Create a Service Instance for the registered service.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: servicecatalog.k8s.io/v1beta1
   kind: ServiceInstance
   metadata:
     name: $INSTANCE_NAME
     namespace: $NAMESPACE
   spec:
     serviceClassExternalName: $EXTERNAL_NAME
   EOF
   ```

4. Check if the Service Instance was created successfully. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance $INSTANCE_NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```
