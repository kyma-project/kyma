---
title: Overview
type: Getting Started
---

This set of Getting Started guides is an end-to-end scenario that will walk you through all major Kyma functionalities and demonstrate its:
- **Integration and connectivity** feature brought in by [Application Connector](https://kyma-project.io/docs/components/application-connector/). With Kyma you can connect external applications and expose their API and events in Kyma.
- **Extensibility** feature provided by [Service Catalog](https://kyma-project.io/docs/components/service-catalog/). You can use its built-in portfolio of external services in your applications.
- **Application runtime** feature supported by [Serverless](https://kyma-project.io/docs/components/serverless/) where you can build microservices and functions to interact with external services and applications to perform a given business logic.

Having these features in mind, we will:

1. Create a microservice that can expose HTTP endpoints and receive a specific type of events from external applications. It has a built-in in-memory storage for storing event details, but it can also work with the external Redis storage.
2. Provision the Redis service through the Service Catalog and use it as an alternative database for the microservice.
3. Connect a mock application as an addon to Kyma. We will use it to send events to the microservice with order delivery details.
4. Create a Function with the logic similar to that of the microservice. We will demonstrate it can also be triggered by order events from the mock application, and save the order details in the external storage.

> **CAUTION:** These tutorials will refer to a sample `orders-service` application deployed on Kyma as a `microservice` to easily distinguish it from the external Commerce mock that represents an external monolithic application connected to Kyma. In Kyma docs, these are referred to as `application` and `Application` respectively.

All guides, whenever possible, demonstrate the steps from both kubectl and Console UI.

## Prerequisites

1. Install:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (1.16 or greater)
- [curl](https://github.com/curl/curl)
- [Kyma](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster]) on a cluster of your choice.

2. Download the `kubeconfig` file with the cluster configuration:

a. Access the Console UI of your Kyma cluster.

b. Click the user icon in the top right corner.

c. Select **Get Kubeconfig** from the drop-down menu to download the configuration file to a selected location on your machine.

d. Open a terminal window.

e. Export the **KUBECONFIG** environment variable to point to the downloaded `kubeconfig`. Run this command:

   ```bash
   export KUBECONFIG={KUBECONFIG_FILE_PATH}
   ```

   >**NOTE:** Drag and drop the `kubeconfig` file in the terminal to easily add the path of the file to the `export KUBECONFIG` command you run.

f. Run `kubectl cluster-info` to check if you are connected to the correct cluster.

<!-- Once the updates Security component in docs is merged, move the above Kubeconfig steps there (existing doc: https://kyma-project.io/docs/components/security/#details-iam-kubeconfig-service-get-the-kubeconfig-file-and-configure-the-cli) and only link to these steps here. -->

## Main actors

Let's introduce the main actors that will lead us through the guides. These are our own examples, mock applications, and experimental services:

- [`orders-service`](https://github.com/kazydek/examples/tree/master/orders-service) is a sample application (microservice) written in Go. It can expose HTTP endpoints used to create and read basic order JSON entities. The service can run with either an in-memory database that is enabled by default or an external, Redis database. On the basis of it, we will show how you can use Kyma to deploy your own microservice, expose it on HTTP endpoints to make it available for external services, and bind it to an external database service (Redis).

- [`orders-function`](https://github.com/kyma-project/examples/blob/order-service/orders-service/deployment/function.yaml) that is an equivalent of the `orders-service` with the ability to retrieve order records and save data in it. Like the microservice, the Function can run with either an in-memory database or a Redis instance.

- [Redis addon](https://github.com/kyma-project/addons/tree/master/addons/redis-0.0.3) is basically a bundle of two Redis services available in two plans: `micro` and `enterprise`. We will connect it to Kyma thanks to Helm Broker and expose it in the Kyma cluster as an addon under Service Catalog. The Redis service represents an open source, in-memory data structure store, used as a database, cache and message broker. For the purpose of these guides, we will use the `micro` plan with the in-memory storage to demonstrate how it can replace the default memory of our microservice.

- [Commerce mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/commerce-mock-0.1.0) is to act as a sample external and monolithic application which we want to extend with Kyma. It is based on the [Varkes](https://github.com/kyma-incubator/varkes) project and is also available in the form of an addon. It will simulate how you can pair an external application with Kyma and expose its APIs and events. In our guides, we will use its `order.deliverysent.v1` event type to trigger both the microservice and the Function.

## Steps

These guides cover the following steps:

1. [Set up a Namespace](#getting-started-create-a-namespace) (`orders-service`) in which you will deploy all resources.
2. Create a microservice by deploying `orders-service` on the cluster in the `orders-service` Namespace.
3. [Expose the microservice](#getting-started-deploy-the-microservice) through the APIRule CR on HTTP endpoints. This way it will be available for other services outside the cluster and you will be able to retrieve data from it and send sample orders. Send a sample order to its endpoint - since it uses the in-memory storage, the order details will be gone after you restart the microservice.
4. [Add the Redis service](#getting-started-add-the-redis-service) as an addon.
5. [Create the Redis service instance](#getting-started-create-a-service-instance-for-the-redis-service) in the Namespace (ServiceInstance CR) so you can bind it with the microservice and Function.
6. [Bind the microservice to the Redis service](#getting-started-bind-the-redis-service-instance-to-the-microservice) by creating ServiceBinding and ServiceBindingUsage CRs. Send a sample order to its endpoint - since it now uses the Redis storage, the order details will not be gone after you restart the microservice.
7. [Connect Commerce mock](#getting-started-connect-an-external-application) as the external application.
8. [Trigger the microservice](#getting-started-trigger-a-microservice-with-an-event) to react to the `order.deliverysent.v1` event from the Commerce mock. Send the event and see if the microservice reacts to it by saving its details in the Redis database.
9. [Create a Function](#getting-started-create-a-function) in the Namespace and repeat the microservice flow:
- [Expose the Function](#getting-started-expose-a-function) through the APIRule CR on HTTP endpoints. This way it will be available for other services outside the cluster.
- [Bind a Function to the Redis service](#getting-started-bind-a-redis-service-instance-to-a-function) by creating ServiceBinding and ServiceBindingUsage CRs.
- [Trigger the Function](#getting-started-trigger-a-function-with-an-event) to react to the `order.deliverysent.v1` event from Commerce mock. Send the event and see if the Function reacts to it by saving its details in the Redis database.

As a result, you get two scenarios of the same flow - a microservice and a Function that are triggered by new order events from Commerce mock and send order data to the Redis database:

![Order flow](./assets/order-flow.svg)
