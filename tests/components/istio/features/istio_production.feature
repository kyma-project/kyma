Feature: Istio is installed
  Istio needs to be installed
  as the core prerequisite.

  Scenario: Istio component installed in production profile has all required pods running
    Given a running Kyma cluster with "production" profile
    When Istio component is installed
    Then there is 2 pod for Pilot
    And there is 3 pod for Ingress gateway
    And Istio pods are available
    And HPA is deployed
