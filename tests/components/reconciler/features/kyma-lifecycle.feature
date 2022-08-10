Feature: Manage Kyma lifecycle on a Kubernetes cluster

  Customers can order an SKR cluster or bring their own
  cluster to retrieve a managed Kyma runtime. The reconciler
  has to manage full lifecycle of a Kyma installation.

  Background:
    Given KCP cluster created
    * SKR cluster created
    * Kyma reconciler installed in KCP cluster

  Scenario: Kyma CR is created with 1 centralized and 1 decentralized module in KCP cluster and Kyma system deployed in SKR cluster
    When Kyma CR with 1 centralized module and 1 decentralized module created in KCP cluster
    Then 1 centralized module CRs created in KCP cluster
    * 1 decentralized module CRs created in SKR cluster
    * 1 manifest CRs created in KCP cluster
    * 1 module deployed in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is deleted in KCP cluster and Kyma system deleted in SKR cluster
    Given Kyma CR with 1 centralized module and 1 decentralized module created in KCP cluster
    When Kyma CR deleted in KCP cluster
    Then module CRs deleted in KCP cluster
    * manifest CRs deleted in KCP cluster
    * module CRs deleted in SKR cluster
    * modules undeployed in SKR cluster

  Scenario: Kyma CR is updated by adding a central module in KCP cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 1 centralized module and 0 decentralized module in KCP cluster
    Then 1 centralized module CR created in KCP cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by deleting a central module in KCP cluster
    Given Kyma CR with 1 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized module and 0 decentralized module in KCP cluster
    Then module CRs deleted in KCP cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by adding a decentralized module in KCP cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized modules and 1 decentralized module in KCP cluster
    Then 1 manifest CR created in KCP cluster
    * 1 decentralized module CR created in SKR cluster
    * 1 module deployed in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by deleting a decentralized module in KCP cluster
    Given Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized modules and 0 decentralized module in KCP cluster
    Then manifest CRs deleted in KCP cluster
    * module CRs deleted in SKR cluster
    * module undeployed in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by adding a central module in SKR cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 1 centralized module and 0 decentralized module in SKR cluster
    Then 1 centralized module CR created in KCP cluster
    * Kyma CR conditions updated in KCP cluster
    * Kyma CR conditions updated in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by deleting a central module in SKR cluster
    Given Kyma CR with 1 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized module and 0 decentralized module in SKR cluster
    Then module CRs deleted in KCP cluster
    * Kyma CR conditions updated in KCP cluster
    * Kyma CR conditions updated in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by adding a decentralized module in SKR cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized modules and 1 decentralized module in SKR cluster
    Then Kyma CR conditions updated in KCP cluster
    * Kyma CR conditions updated in SKR cluster
    * 1 manifest CR created in KCP cluster
    * 1 decentralized module CR created in SKR cluster
    * 1 module deployed in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is updated by deleting a decentralized module in SKR cluster
    Given Kyma CR with 0 centralized modules and 1 decentralized modules created in KCP cluster
    When Kyma CR updated by setting 0 centralized modules and 0 decentralized module in SKR cluster
    Then  Kyma CR conditions updated in KCP cluster
    * Kyma CR conditions updated in SKR cluster
    * manifest CRs deleted in KCP cluster
    * module CRs deleted in SKR cluster
    * module undeployed in SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is deleted in SKR cluster and recovered from KCP cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR deleted in SKR cluster
    Then Kyma CR copied from KCP to SKR cluster
    * Kyma CR in state READY within 10sec

  Scenario: Kyma CR is update with invalid change in SKR cluster
    Given Kyma CR with 0 centralized modules and 0 decentralized modules created in KCP cluster
    When Kyma CR updated with invalid change in SKR cluster
    Then Kyma CR contains event with warning
    * Kyma CR in state ERROR within 10sec
    * Validating webhook logs error