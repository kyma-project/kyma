Feature: Istio reconciler works as expected
  Scenario: A httpbin deployment without sidecar gets restarted and injected with sidecar
    Given Istio component is installed
    And "sidecar-recovery" namespace exists
    And "sidecar-recovery" namespace is labeled with "istio-injection" "enabled"
    And Httpbin deployment is created in "sidecar-recovery" namespace
    And Httpbin deployment is deployed and ready in "sidecar-recovery" namespace
    And there "should" be Istio sidecar in httpbin pod in "sidecar-recovery" namespace
    And a reconcilation takes place
    And istioctl install takes place
    And the httpbin deployment in "sidecar-recovery" namespace gets restarted until there is no sidecar
    And the reconcilation is aborted
    And there is 1 Httpbin deployment in "sidecar-recovery" namespace
    And there "shouldn't" be Istio sidecar in httpbin pod in "sidecar-recovery" namespace
    When a reconcilation takes place
    Then reconciler restarts the faulty deployment
    And Httpbin deployment is deployed and ready in "kyma-system" namespace
    And there "should" be Istio sidecar in httpbin pod in "sidecar-recovery" namespace
    And "sidecar-recovery" namespace is deleted
