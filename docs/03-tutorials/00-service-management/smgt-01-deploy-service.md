---
title: Deploy an SAP BTP service in your Kyma cluster
---

This tutorial describes how you can deploy a simple SAP BTP audit log service in your Kyma cluster using the [SAP BTP service operator](https://github.com/SAP/sap-btp-service-operator).

## Prerequisites

- [Kyma cluster](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/02-install-kyma/) running on Kubernetes v1.26 or higher
- SAP BTP [Global Account](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/d61c2819034b48e68145c45c36acba6e.html?locale=en-US) and [Subaccount](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/55d0b6d8b96846b8ae93b85194df0944.html?locale=en-US)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) v1.26 or higher
- [helm](https://helm.sh/) v3.0 or higher
- [jq](https://stedolan.github.io/jq/download/)


## Steps

1. Create a Namespace and install [cert-manager](https://cert-manager.io/docs/) in it. The SAP BTP operator requires cert-manager to work properly. You can skip this step if you have cert-manager already installed. Run:

    ```bash
    kubectl create ns cert-manager
    kubectl label namespace cert-manager istio-injection=disabled
    kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
    echo "Wait two minutes. For details, read: https://cert-manager.io/docs/concepts/webhook/#webhook-connection-problems-shortly-after-cert-manager-installation"
    sleep 120
    ```

    >**CAUTION:** There's a known issue with the [webhook connection](https://cert-manager.io/docs/concepts/webhook/#webhook-connection-problems-shortly-after-cert-manager-installation) shortly after cert-manager installation, and with the BTP operator webhook. If you see the `failed calling webhook` error after running this and/or the next command, wait a moment and repeat the operation.

2. Obtain the access credentials for the SAP BTP service operator as described in step 2 of the [SAP BTP operator setup](https://github.com/SAP/sap-btp-service-operator#setup). Then, save the credentials to the `creds.json` file.

3. Create a Namespace and install the SAP BTP service operator in it:

    ```bash
    kubectl create ns sap-btp-operator
    kubectl label namespace sap-btp-operator istio-injection=disabled
    helm repo add sap-btp-operator https://sap.github.io/sap-btp-service-operator
    helm upgrade --install btp-operator sap-btp-operator/sap-btp-operator --create-namespace --namespace=sap-btp-operator --set manager.secret.clientid="$(jq --raw-output '.clientid' creds.json)" --set manager.secret.clientsecret="$(jq --raw-output '.clientsecret' creds.json)" --set manager.secret.sm_url="$(jq --raw-output '.sm_url' creds.json)" --set manager.secret.tokenurl="$(jq --raw-output '.url' creds.json)"

    echo "Wait 30 seconds to make btp-operator webhook ready"
    sleep 30
    ```

4. Create a Service Instance:

    ```yaml
    kubectl create -f - <<EOF
    apiVersion: services.cloud.sap.com/v1alpha1
    kind: ServiceInstance
    metadata:
      name: btp-audit-log-instance
      namespace: default
    spec:
      serviceOfferingName: auditlog-api
      servicePlanName: default
      externalName: btp-audit-log-instance
    EOF
    ```

    >**TIP:** You can find values for the **serviceOfferingName** and **servicePlanName** parameters in the Service Marketplace of the SAP BTP Cockpit. Click on the service's tile and find **name** and **Plan** respectively. The value of the **externalName** parameter must be unique.

5. To see the output, run:

    ```bash
    kubectl get serviceinstances.services.cloud.sap.com btp-audit-log-instance -o yaml
    ```

    You can see the status `created` and the message `ServiceInstance provisioned successfully`.

6. Create a Service Binding:

    ```yaml
    kubectl create -f - <<EOF
    apiVersion: services.cloud.sap.com/v1alpha1
    kind: ServiceBinding
    metadata:
      name: btp-audit-log-binding
      namespace: default
    spec:
      serviceInstanceName: btp-audit-log-instance
      externalName: btp-audit-log-binding
      secretName: btp-audit-log-binding
    EOF
    ```

7. To see the output, run:

    ```bash
    kubectl get servicebindings.services.cloud.sap.com btp-audit-log-binding -o yaml
    ```

    You can see the status `created` and the message `ServiceBinding provisioned successfully`.

8. You can now use a given service in your Kyma cluster. To see credentials, run:

    ```bash
    kubectl get secret btp-audit-log-binding -o yaml
    ```

9. Clean up your resources:

    ```bash
    kubectl delete servicebindings.services.cloud.sap.com btp-audit-log-binding
    kubectl delete serviceinstances.services.cloud.sap.com btp-audit-log-instance
    helm delete btp-operator -n sap-btp-operator
    kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
    kubectl delete ns cert-manager
    kubectl delete ns sap-btp-operator
    ```

>**TIP:** You can use Kyma Dashboard to create and manage resources such as ServiceInstances and ServiceBindings. To do so, navigate to your Namespace view and go to the **Service Management** tab in the left navigation. Still, you need to obtain service details, such as service name and plan, from the BTP Cockpit.
