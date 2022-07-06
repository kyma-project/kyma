Feature: Sidecar injection is enabled in "kyma-system" by default

  Scenario: Kyma-system does not contain pods with sidecar
    Given Istio component is installed
    Then there should be pods with istio sidecar in "kyma-system" namespace
    Given Httpbin deployment is created in "kyma-system"
    When Httpbin deployment should be deployed and ready in "kyma-system" namespace
    Then there should be istio sidecar in httpbin pod in "kyma-system" namespace
    And Httpbin deployment is deleted from "kyma-system"
