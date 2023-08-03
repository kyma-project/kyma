---
title: What is Service Management in Kyma?
---

Service Management in Kyma uses service operators provided by third-party services. You can find the base of available service operators in the [OperatorHub.io](https://operatorhub.io/).

Service Management in Kyma uses service operators. A service operator is a piece of software that provides a set of all necessary resources (for example, CustomResourceDefinitions and controllers) needed to provision third-party services in your Kubernetes cluster. With service operators, you can manage your Kubernetes-native applications without worrying about technicalities behind operational activities in your cluster, such as installation, configuration, or modification.

Many third-party service providers offer their own operators to help you provision their services. These are the examples of operators that you can install on your Kyma cluster:
- [SAP BTP](https://github.com/SAP/sap-btp-service-operator)
- [GCP](https://cloud.google.com/config-connector/docs/how-to/getting-started)
- [Azure](https://github.com/Azure/azure-service-operator)
- [AWS](https://github.com/aws-controllers-k8s/community)

You can find the base of all service operators in the [OperatorHub.io](https://operatorhub.io/).

## Useful links

If you're interested in learning more about the Service Management area, check out these links to:

- [Deploy an SAP BTP service in your Kyma cluster](../../03-tutorials/00-service-management/smgt-01-deploy-service.md)
