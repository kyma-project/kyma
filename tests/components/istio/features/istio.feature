Feature: istio is installed
  Istio needs to be installed
  as the core prerequisite.

  Scenario: istio is installed with evaluation profile
    Given installed istio
    Then there should be at least 2 pods
    And they should be available for at least 3 seconds
