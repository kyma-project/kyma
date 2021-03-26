---
title: Service Brokers
type: Overview
---

A Service Broker is a server compatible with the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md) specification. Each Service Broker registered in Kyma presents the services it offers to the Service Catalog and manages their lifecycle.

The Service Catalog lists all services that Service Brokers offer. Use the Service Brokers to:
* Provision and deprovision an instance of a service.
* Create and delete a ServiceBinding to link a ServiceInstance to an application.

Each of the Service Brokers available in Kyma performs these operations in a different way. See the documentation of a given Service Broker to learn how it operates. The Service Catalog provided by Kyma is currently integrated with the following Service Brokers:

* [Application Broker](/components/application-connector#architecture-application-connector-components-application-broker)
* [Helm Broker](/components/helm-broker/#overview-overview)

You can also use Helm Broker addons to install these third-party brokers:

<div tabs name="brokers" group="brokers">
  <details>
  <summary label="azure-service-broker">
  Azure Service Broker
  </summary>

  The Microsoft Azure Service Broker is an open-source, [Open Service Broker](https://www.openservicebrokerapi.org/)-compatible API server that provisions managed services in the Microsoft Azure public cloud. Kyma provides the Namespace-scoped Azure Service Broker. In each Namespace, you can configure the Azure Service Broker against different subscriptions. Install the Azure Service Broker by provisioning the **Azure Service Broker** class provided by the Helm Broker.

  ![azure broker class](./assets/azure-service-broker-class.png)

  Once you provision the **Azure Service Broker** class, the Azure Service Broker classes are available in the Service Catalog view in a given Namespace.
  The Azure Service Broker provides these ServiceClasses to use with the Service Catalog:

  * Azure SQL Database
  * Azure Database for MySQL
  * Azure Redis Cache
  * Azure Application Insights
  * Azure CosmosDB
  * Azure Event Hubs
  * Azure IoT Hub
  * Azure Key Vault
  * Azure SQL Database
  * Azure SQL Database Failover Group
  * Azure Service Bus
  * Azure Storage
  * Azure Text Analytics

  See the details of each ServiceClass and its specification in the Service Catalog UI.

  </details>
  <details>
  <summary label="aws-service-broker">
  AWS Service Broker
  </summary>

  The AWS Service Broker is an open-source, [Open Service Broker](https://www.openservicebrokerapi.org/)-compatible API server that provisions managed services in the AWS public cloud. Kyma provides the Namespace-scoped AWS Service Broker. In each Namespace, you can configure the AWS Service Broker against different subscriptions. Install the AWS Service Broker by provisioning the **AWS Service Broker** class provided by the Helm Broker.

  ![aws broker class](./assets/aws-class.png)

  Once you provision the **AWS Service Broker** class, the AWS Service Broker classes are available in the Service Catalog view in a given Namespace.
  The AWS Service Broker provides these ServiceClasses to use with the Service Catalog:

  * Amazon Athena
  * Amazon EMR
  * Amazon Kinesis
  * Amazon RDS for MariaDB
  * Amazon RDS for PostgreSQL
  * Amazon Translate
  * Amazon KMS
  * Amazon Rekognition
  * Amazon SNS
  * Amazon DynamoDB
  * Amazon Redshift
  * Amazon SQS
  * Amazon Polly
  * Amazon RDS for MySQL
  * Amazon S3
  * Amazon Lex
  * Amazon Route53
  * Amazon ElasticCache
  * Amazon ElasticSearch
  * Amazon DocumentDB
  * Amazon RDS for PostgreSQL
  * Amazon RDS for Oracle
  * Amazon RDS for Mssql
  * Amazon Aurora PostgreSQL
  * Amazon Aurora MySQL

  See the [documentation for each ServiceClass](https://github.com/awslabs/aws-servicebroker/tree/v1.0.0/templates). You can also see the details and specification of each ServiceClass in the Service Catalog UI, after provisioning a given class.

  >**NOTE:** Kyma uses the AWS Service Broker open-source project. To ensure the best performance and stability of the product, Kyma uses a version of the AWS Service Broker that precedes the newest version released by Amazon.

  </details>
  <details>
  <summary label="gcp-service-broker">
  GCP Service Broker
  </summary>

  The GCP Service Broker is an open-source, [Open Service Broker](https://www.openservicebrokerapi.org/)-compatible API server that provisions managed services in the Google Cloud Platform public cloud. Kyma provides the Namespace-scoped GCP Service Broker. In each Namespace, you can configure the GCP Service Broker against different subscriptions. Install the GCP Service Broker by provisioning the **GCP Service Broker** class provided by the Helm Broker.

  ![gcp broker class](./assets/gcp-class.png)

  Once you provision the **GCP Service Broker** class, the GCP Service Broker classes are available in the **Service Catalog** view in the given Namespace.
  The GCP Service Broker provides these ServiceClasses to use with the Service Catalog:

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

  See the details of each ServiceClass and its specification in the Service Catalog UI.

   </details>
</div>


To get the addons that the Helm Broker provides, go to the [`addons`](https://github.com/kyma-project/addons) repository. To build your own Service Broker, follow the [Open Service Broker API specification](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md). For details on how to [register a sample Service Broker in the Service Catalog](#tutorials-register-a-broker-in-the-service-catalog), go to the **Tutorials** section.

>**NOTE:** The Service Catalog has the Istio sidecar injected. To enable the communication between the Service Catalog and Service Brokers, either inject Istio sidecar into all brokers or disable mutual TLS authentication.
