---
title: Google Cloud Platform Broker
type: Overview
---

>**NOTE:** The Google Cloud Platform (GCP) Service Broker is in the experimental phase.

Google Cloud Platform (GCP) Service Broker is an implementation of the [Open Service Broker](https://www.openservicebrokerapi.org/) (OSB) API hosted on GCP. It simplifies the delivery of GCP services to applications that run on Kyma. By creating GCP resources and managing their corresponding permissions, Service Broker makes it easy to consume GCP services from within a Kubernetes cluster.

Kyma provides Namespace-scoped GCP Service Broker. In each Namespace, you can configure the GCP Broker against different Google Cloud Platforms. Install the GCP Service Broker by provisioning the **Google Cloud Platform Service Broker Provider** class provided by the Helm Broker. 

Configure the Helm Broker to enable the **Google Cloud Platform Service Broker Provider** class:
```bash
kubectl set env -n kyma-system deployment/core-helm-broker -e APP_REPOSITORY_URLS="https://github.com/kyma-project/bundles/releases/download/latest/index.yaml"
```

![Service Catalog view without GCP Classes](assets/003-gcp-provider-class.png)

Once you provision the **Google Cloud Platform Service Broker Provider** class, the GCP Service Broker classes are available in the Service Catalog view in a given Namespace.

![Service Catalog view without GCP Classes](assets/003-gcp-broker-classes.png)

For more information about provisiong the **Google Cloud Platform Service Broker Provider** class, go to the service class overview in the Service Catalog UI.

>**NOTE:** Google does not provide the ability to specify the broker's version so the most up-to-date version of the GCP Broker is installed.
