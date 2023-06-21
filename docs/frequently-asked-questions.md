---
title: FAQ
---

### Why should I use Kyma? What are the advantages?
Kyma is built upon leading cloud-native, open-source projects and open standards, such as Istio, NATS, Cloud Events, Open Telemetry, and Prometheus. We created an opinionated set of modules you can easily enable in your Kubernetes cluster to speed up cloud application development and operations. With Kyma, you save the time to pick the right tools and the effort to keep them secure and up to date. 

### Is Kyma just another platform?

No, Kyma is not intended to be a full-scale platform. It is a Kubernetes-based application runtime with several extensions. Those extensions make Kyma more attractive for developers who want to focus more on business logic and limit investment in technical services and infrastructure. Kyma is part of SAP Business Technology Platform and offers easy integration with BTP services and other SAP systems.

### What is the difference between open source Kyma project and SPA BTP Kyma Runtime?

SAP BTP Kyma Runtime is a bundle of a Kubernetes cluster powered by Gardener and Kyma modules provided as a managed service. All the components are regularly updated, and the availability is monitored and guaranteed (SLA). Managed Kyma runtime is also preconfigured to easily connect to other SAP services and systems. Using Kyma Runtime, you can face some limitations in configuring Kyma components as some settings are managed centrally and overwrite user changes, but still, you get the admin access to the cluster, that is the cluster-admin role.
When using Kyma open-source components, you have more control and flexibility over installation, configuration, and upgrade processes but more responsibilities related to operations.

### Can I combine Kyma modules with other Kubernetes tools?

Yes. Kyma is opinionated but open. You can pick the modules you need from Kyma and complement them with other tools. 

### Kyma is a part of the commercial product (SAP BTP Kyma Runtime). What is the chance that Kyma will become a closed project?

Kyma has been open source since 2018 and part of SAP BTP since 2019. We believe that openness and vendor independence is a valuable proposition. We also believe that offering an open-source project as a commercial product only benefits both parties. Open-source users get the confidence that the project won't be abandoned anytime soon, and customers see the quality and technical details of the product. Apart from that, SAP strongly supports the open-source community. For more information, visit [SAP Open Source](https://community.sap.com/topics/open-source).
