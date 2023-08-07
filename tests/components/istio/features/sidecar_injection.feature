Feature: Istio sidecar injection works properly in target namespace

  Scenario: Namespace with istio-injection=disabled label does not contain pods with istio sidecar
    Given Istio component is installed
    When "sidecar-disable" namespace exists
    And "sidecar-disable" namespace is labeled with "istio-injection" "disabled"
    And Httpbin deployment is created in "sidecar-disable" namespace
    And Httpbin deployment is deployed and ready in "sidecar-disable" namespace
    Then there should be no pods with Istio sidecar in "sidecar-disable" namespace
    And "sidecar-disable" namespace is deleted

  Scenario: Namespace with istio-injection=enabled label does contain pods with istio sidecar
    Given Istio component is installed
    When "sidecar-enable" namespace exists
    And "sidecar-enable" namespace is labeled with "istio-injection" "enabled"
    And Httpbin deployment is created in "sidecar-enable" namespace
    And Httpbin deployment is deployed and ready in "sidecar-enable" namespace
    Then Pods in namespace "sidecar-enabled" should have proxy sidecar
    And "sidecar-enable" namespace is deleted

  Scenario: Kyma-system namespace contains pods with sidecar
    Given Istio component is installed
    Then there should be some pods with Istio sidecar in "kyma-system" namespace
    Given Httpbin deployment is created in "kyma-system" namespace
    When Httpbin deployment is deployed and ready in "kyma-system" namespace
    Then there "should" be Istio sidecar in httpbin pod in "kyma-system" namespace
    And Httpbin deployment is deleted from "kyma-system" namespace

  Scenario: Kube-system namespace does not contain pods with sidecar
    Given Istio component is installed
    Then there should be no pods with Istio sidecar in "kube-system" namespace
    Given Httpbin deployment is created in "kube-system" namespace
    When Httpbin deployment is deployed and ready in "kube-system" namespace
    Then there should be no pods with Istio sidecar in "kube-system" namespace
    And Httpbin deployment is deleted from "kube-system" namespace
