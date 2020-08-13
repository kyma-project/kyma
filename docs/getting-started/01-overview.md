---
title: How to start
type: Getting Started
---

This set of Getting Started guides aims to demonstrate the basic functionalities offered by Kyma its:
- **Integration and connectivity** feature brought in by [Application Connector](https://kyma-project.io/docs/components/application-connector/). With Kyma you can connect external applications and expose their API and events in Kyma.
- **Extensibility** feature provided by [Service Catalog](https://kyma-project.io/docs/components/service-catalog/). You can use its built-in portfolio of external services in your applications.
- **Application runtime** feature supported by [Serverless](https://kyma-project.io/docs/components/serverless/) where you can build microservices and functions to interact with external services and applications to perform a given business logic.

Having these features in mind, we will:
1. Connect a mock application as an addon to Kyma and use one of its events sent whenever an order is created in an application by a user.
2. Provision the Redis service available through the Service Catalog to act as an external database in which the order information can be stored.
3. Create a microservice and a function to combine the previous two pieces of the puzzle. You can use both of them in Kyma to perform a certain business logic, integrate external applications and services and create meaningful flows. In our guides, you will see both of them used in the same flow - thanks to their logic, both the microservice and the function react to the event sent from the mock application and store the order data in the attached Redis database.

> **CAUTION:** These tutorials will refer to a sample `orders-service` application deployed on Kyma as a `microservice` to easily distinguish it from the external Marketing mock that represents an external monolithic application connected to Kyma. In Kyma docs, these are referred to as `application` and `Application` respectively.

All guides, whenever possible, will demonstrate the steps to perform both from kubectl and Console UI.

## Prerequisites

These guides show what you can do with Kyma running on a cluster of your choice. Before you start, follow the steps in the [installation tutorials](https://kyma-project.io/docs/#installation-install-kyma-on-a-cluster]) to get your Kyma cluster up and running.

## Main actors

Let's introduce the main actors that will lead us through the guides. These are our own examples, mock applications, and experimental services:

- [Orders service](https://github.com/kazydek/examples/tree/master/orders-service) is a sample application (microservice) written in Go. It can expose HTTP endpoints used to create and read basic order JSON entities. The service can run with either an in-memory database that is enabled by default or an external, Redis database. On the basis of it, we will show how you can use Kyma to deploy your own microservice, expose it on HTTP endpoints to make it available for external services, and bind it to an actual database service (Redis).

- Function - _TODO_

- [Redis addon](https://github.com/kyma-project/addons/tree/master/addons/redis-0.0.3) is basically a bundle of two Redis services available in two plans: `micro` and `enterprise`. We will connect it to Kyma thanks to Helm Broker and expose it in the Kyma cluster as an addon under Service Catalog. The Redis service represents an open source, in-memory data structure store, used as a database, cache and message broker. For the purpose of these guides, we will use the `micro` plan with the in-memory storage to demonstrate how it can replace the default memory of our microservice.

- [Marketing mock](https://github.com/SAP-samples/xf-addons/tree/master/addons/marketing-mock-0.1.0) is to act as a sample external and monolithic application which we want to extend with Kyma. It is based on the [Varkes](https://github.com/kyma-incubator/varkes) project and is also available in the form of an addon. It will simulate how you can pair an external application with Kyma and expose its APIs and Events. In our guides, we will use its **bo.interaction.created** event.

## Steps

The guides cover these steps:

1. Connect the external SAP Marketing Cloud - Mock application.
2. Add a Redis service as an addon.
3. Create a ServiceInstance CR for the Redis service so you can bind it later with your microservice and Function.
4. Create a microservice by deploying `orders-service` on the cluster.
5. Expose the microservice through the APIRule CR on HTTP endpoints. This way it will be available for other services outside the cluster.
6. Bind a microservice to the Redis service by creating ServiceBinding and ServiceBindingUsage CRs.
7. Trigger your microservice to react to the **bo.interaction.created** event from the mock application. Send the event and see if the microservice reacts to it by saving its details in the Redis database.
8. Create a Function on the cluster.
9. Expose the Function through the APIRule CR on HTTP endpoints. This way it will be available for other services outside the cluster.
10. Bind a Function to the Redis service by creating ServiceBinding and ServiceBindingUsage CRs.
11. Trigger your Function to react to the **bo.interaction.created** event from the mock application. Send the event and see if the Function reacts to it by saving its details in the Redis database.

As a result, you get two scenarios of the same flow - a microservice and a Function that are triggered by new order events from the Marketing mock and send order data to the Redis database:

![Order flow](./assets/order-flow.svg)
