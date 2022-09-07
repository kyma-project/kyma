Feature: Istio reconciler works as expected
  Scenario: A httpbin deployment without sidecar gets restarted and injected with sidecar
    Given Istio component is installed
    And "sidecar-recovery" namespace exists
    And "sidecar-recovery" namespace is labeled with "istio-injection" "enabled"
    And Httpbin deployment is created in "sidecar-recovery" namespace
    And there "should" be Istio sidecar in httpbin pod in "sidecar-recovery" namespace
    When a reconcilation takes place
    And istioctl install takes place
    And the httpbin deployment in "sidecar-recovery" namespace gets restarted until there is no sidecar
    Then reconciler restarts the faulty deployment
    And there "should" be Istio sidecar in httpbin pod in "sidecar-recovery" namespace
    And "sidecar-recovery" namespace is deleted
