Feature: kube-system sidecar injection disabled
  Istio needs to be installed
  as the core prerequisite.

  Scenario: Kube-system does not contain pods with sidecar
    Given Istio component is installed
    Then there should be no pods with istio sidecar in kube-system namespace 