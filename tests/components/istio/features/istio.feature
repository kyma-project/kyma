Feature: Istio is installed
  Istio needs to be installed
  as the core prerequisite.

  Scenario: Istio component installed in evaluation profile has all required pods running
    Given a running Kyma cluster with "evaluation" profile
    When Istio component is installed
    Then there is 1 pod for Pilot
    And there is 1 pod for Ingress gateway
    And Istio pods are available
    And HPA is not deployed
