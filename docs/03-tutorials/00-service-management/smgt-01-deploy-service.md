---
title: Deploy an SAP BTP service in your Kyma cluster
---

This tutorial describes how you can deploy an SAP BTP service in your Kyma cluster using the SAP BTP service operator.

## Prerequisites

- [Kyma cluster](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/02-install-kyma/) running on Kubernetes version 1.19 or higher
- SAP BTP [Global Account](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/d61c2819034b48e68145c45c36acba6e.html?locale=en-US) and [Subaccount](https://help.sap.com/products/BTP/65de2977205c403bbc107264b8eccf4b/55d0b6d8b96846b8ae93b85194df0944.html?locale=en-US)
- Service Management Control ([SMCTL](https://help.sap.com/viewer/09cc82baadc542a688176dce601398de/Cloud/en-US/0107f3f8c1954a4e96802f556fc807e3.html)) command line interface
<!---
^ Do we need this?
-->
- [kubectl](https://kubernetes.io/docs/tasks/tools/) v1.17 or higher
- [helm](https://helm.sh/) v3.0 or higher

## Steps

1. Create a Namespace for SAP BTP Operator:
```
kubectl create ns sap-btp-operator
```

2. For the BTP service operator to work, you must [disable Istio sidecar proxy injection](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/smsh-01-istio-disable-sidecar-injection#documentation-content) that is enabled on the Kyma clusters by default. To do so, run this command:
```
kubectl label namespace sap-btp-operator istio-injection=disabled
```

<!---
2. Alternative if the previous option doesn't work:

kubectl patch deployment -ncert-manager cert-manager-webhook --type=json -p '[{ "op": "add", "path": "/spec/template/metadata/annotations", "value": {"sidecar.istio.io/inject": "false"} }]'
-->

3. Install and set up the [SAP BTP service operator](https://github.com/SAP/sap-btp-service-operator) in your Kyma cluster. Use the option without cert-manager.

4. Create a Service Instance:

```yaml
kubectl create -f - <<EOF
apiVersion: services.cloud.sap.com/v1alpha1
kind: ServiceInstance
metadata:
  name: btp-auditlog-2
  namespace: default
spec:
  serviceOfferingName: auditlog-api
  servicePlanName: default
  externalName: btp-auditlog-2
EOF
```

>**TIP:** You can find values for the **serviceOfferingName** and **servicePlanName** parameters in Service Marketplace of SAP BTP Cockpit. Click on the service's tile and find **name** and **Plan** respectively. The value of the **externalName** parameter must be unique.

5. To see the output, run:

```
kubectl get serviceinstance.btp<api group> btp-auditlog-2 -o yaml   
```

You can see the status "created" and the message "ServiceInstance provisioned successfully".

6. Create a Service Binding:

```yaml
kubectl create -f - <<EOF
apiVersion: services.cloud.sap.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding
  namespace: default
spec:
  serviceInstanceName: btp-auditlog-2
  externalName: btp-operator-auditlog-api-binding
  secretName: btp-operator-auditlog-api
EOF
```

7. To see the output, run:

```
kubectl get serviceinstance.btp<api group> btp-auditlog-2 -o yaml   
```

You can see the status "created" and the message "ServiceBinding provisioned successfully".

8. You can now use a given service in your Kyma cluster.

<!---
You can use Kyma Dashboard to create and manage resources such as Service Instance and Service Binding. In the left navigation, go to the Service Management > BTP . Still, you need to acquire service details from BTP Cockpit.
-->
