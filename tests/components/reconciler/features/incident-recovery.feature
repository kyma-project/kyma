Feature: Manage Kyma lifecycle on a Kubernetes cluster

  Customers can apply changes in their Kubernetes clusters
  which can lead to misbehaving or interrupted reconcliation processes.
  The reconciler has to be able to react on expected situations
  accordingly and apply self-healing actions.

  Background:
    Given KCP cluster created
    * SKR cluster created
    * Kyma reconciler installed in KCP cluster

  Scenario: Customer deletes Kyma watcher in SKR cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When watcher webhook deleted
    Then watcher webhook created within 300sec
    * SKR cluster reconciled within 360sec

  Scenario: Customer deletes module operator in SKR cluster
    Given Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    When module operator deleted in SKR cluster
    Then module operator created within 300sec in SKR cluster

  Scenario: Customer deletes module CRD in SKR cluster
    Given Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    When module CRD deleted in SKR cluster
    Then module CRD created within 300sec in SKR cluster
