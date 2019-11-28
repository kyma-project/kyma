---
title: GCP Service Broker
type: Service Brokers
---

The GCP Service Broker is an open-source, [Open Service Broker](https://www.openservicebrokerapi.org/)-compatible API server that provisions managed services in the Google Cloud Platform public cloud. Kyma provides Namespace-scoped GCP Service Broker. In each Namespace, you can configure the GCP Service Broker against different subscriptions. Install the GCP Service Broker by provisioning the **GCP Service Broker** class provided by the Helm Broker.

![gcp broker class](./assets/gcp-class.png)

Once you provision the **GCP Service Broker** class, the GCP Service Broker classes are available in the **Service Catalog** view in a given Namespace.
The GCP Service Broker provides these Service Classes to use with the Service Catalog:

* Google BigQuery
* Google Bigtable
* Google CloudSQL for MySQL
* Google CloudSQL for PostgreSQL
* Google Cloud Dataflow
* Google Cloud Datastore
* Google Cloud Dialogflow
* Google Cloud Filestore
* Google Cloud Firestore
* Google Cloud Memorystore for Redis API
* Google Machine Learning APIs
* Google PubSub
* Google Spanner
* Stackdriver Debugger
* Stackdriver Monitoring
* Stackdriver Profiler
* Stackdriver Trace
* Google Cloud Storage

See the details of each Service Class and its specification in the Service Catalog UI.
For more information about the Service Brokers, see [this](#service-brokers-service-brokers) document.
