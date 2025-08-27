# Working with Multiple Subaccounts

With the SAP BTP Operator module, you can create configurations for several subaccounts in a single Kyma cluster.

## Context

By default, a Kyma cluster is associated with one subaccount. Consequently, any service instance created within any namespace is provisioned in the associated subaccount. See [Preconfigured Credentials and Access](03-10-preconfigured-secret.md). However, with SAP BTP Operator, you can create configurations in a single Kyma cluster that are applied to several subaccounts.
To apply the multitenancy feature, choose the method that suits your needs and application architecture. You have the following options:

* [Namespace-level mapping](03-22-namespace-level-mapping.md): Connect namespaces to separate subaccounts by configuring dedicated credentials for each namespace.
* [Instance-level mapping](03-21-instance-level-mapping.md): Define a specific subaccount for each service instance, regardless of the namespace context.

Regardless of the method, you must create Secrets managed in the `kyma-system` namespace.

### Secrets Precedence

SAP BTP Operator searches for the credentials in the following order:

1. Explicit Secret defined in a service instance
2. Managed namespace Secret assigned to a given namespace
3. Managed namespace default Secret

![Secrets precedence](../assets/secrets_precedence_4.drawio.svg) 

## Procedure

* To connect a namespace to a specific subaccount, see [Namespace-Level Mapping](03-22-namespace-level-mapping.md).
* To deploy service instances belonging to different subaccounts within the same namespace, see [Instance-Level Mapping](03-21-instance-level-mapping.md).
