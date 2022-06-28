Feature: Sidecar injection is disabled in kube-system

  Scenario: Kube-system does not contain pods with sidecar
    Given Istio component is installed
    Then there should be no pods with istio sidecar in kube-system namespace
    
    Given Httpbin deployment is created in kube-system
    When Httpbin deployment should be deployed and ready
    Then there should be no pods with istio sidecar in kube-system namespace
    And Httpbin deployment is deleted from kube-system
