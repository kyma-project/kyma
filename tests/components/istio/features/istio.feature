Feature: Istio is installed
  Istio needs to be installed
  as the core prerequisite.

  Scenario: Istio is installed with evaluation profile
    Given installed Istio "evaluation" profile
    Then there should be 1 pod for pilot and 1 pod for ingress gateway
    And HPA should not be deployed
    And Istio pods should be available for at least 3 seconds
