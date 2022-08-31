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

  Scenario: Istio component installed in evaluation profile has resources configured correctly
    Given a running Kyma cluster with "evaluation" profile
    When Istio component is installed
    Then "proxy" has "requests" set to cpu - "10m" and memory - "32Mi"
    And "proxy" has "limits" set to cpu - "250m" and memory - "254Mi"
    And "ingress-gateway" has "requests" set to cpu - "10m" and memory - "32Mi"
    And "ingress-gateway" has "limits" set to cpu - "500m" and memory - "128Mi"
    And "proxy_init" has "requests" set to cpu - "10m" and memory - "10Mi"
    And "proxy_init" has "limits" set to cpu - "100m" and memory - "50Mi"
    And "pilot" has "requests" set to cpu - "10m" and memory - "128Mi"
    And "pilot" has "limits" set to cpu - "250m" and memory - "384Mi"
    And "egress-gateway" has "requests" set to cpu - "10m" and memory - "120Mi"
    And "egress-gateway" has "limits" set to cpu - "2000m" and memory - "1024Mi"
