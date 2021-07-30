---
title: What is Service Management in Kyma?
---

Service Management in Kyma is based on service operators. A service operator is a piece of software that helps you provision third-party services on your Kubernetes cluster by providing a set of all necessary resources (such as CustomResourceDefinitions and controllers) needed to provision those services. With service operators, you can manage your Kubernetes-native applications without worrying about technicalities behind operational activities in your cluster, such as installation, configuration, or modification.

Many third-party service providers offer their own operators to help you provision their services. These are the operators provided by three main hyperscale cloud providers that you can install on your Kyma cluster:
- [GCP](https://cloud.google.com/config-connector/docs/how-to/getting-started)
- [Azure](https://github.com/Azure/azure-service-operator)
- [AWS](https://github.com/aws-controllers-k8s/community)

You can find the base of all service operators in the [OperatorHub.io](https://operatorhub.io/).
