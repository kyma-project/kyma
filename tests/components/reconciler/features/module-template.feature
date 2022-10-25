Feature: Manage Kyma lifecycle on a Kubernetes cluster

  Customers can order an SKR cluster or bring their own
  cluster to retrieve a managed Kyma runtime. The reconciler
  has to manage full lifecycle of a Kyma installation.

  Background:
    Given KCP cluster created
    * SKR cluster created
    * Kyma reconciler installed in KCP cluster

  Scenario: Centralized module template is updated
    Given Kyma CR with 1 centralized modules and 0 decentralized modules created in KCP cluster
    * Kyma CR with 1 centralized modules and 0 decentralized modules created in KCP cluster
    When centralized module template updated
    Then centralized module CRs updated in KCP cluster

  Scenario: Decentralized module template is updated
    Given Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    * Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    When decentralized module template updated
    Then decentralized module CRs updated in SKR cluster