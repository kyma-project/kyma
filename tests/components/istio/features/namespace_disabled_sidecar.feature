Feature: Istio sidecar injection disabled in target namespace

  Scenario: Namespace with istio-injection=disabled label does not contain pods with istio sidecar
    Given Istio component is installed
    When "test-sidecars" namespace exists
    And "test-sidecars" namespace is labeled with "istio-injection" "disabled"
    And Httpbin is deployed in "test-sidecars" namespace
    And Httpbin deployment is deployed and ready in "test-sidecars" namespace
    Then there should be no pods with istio sidecar in "test-sidecars" namespace
    And "test-sidecars" namespace is deleted
