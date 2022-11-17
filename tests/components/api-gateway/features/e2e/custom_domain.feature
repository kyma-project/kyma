Feature: Exposing endpoint with a custom domain
  Scenario: Register new custom domain on the cluster
    Given there is a "google-credentials" DNS cloud credentials secret in "default" namespace
    Then there is an "istio-ingressgateway" service in "istio-system" namespace
    And create custom domain resources
    And ensure that DNS record is ready

  Scenario: Calling an unsecured API endpoint with custom domain
    Given there is an unsecured endpoint
    Then calling the endpoint without a token should result in status between 200 and 299
    And calling the endpoint with any token should result in status between 200 and 299
 
  Scenario: Calling a secured API with OAuth2 with custom domain
    Given endpoint is secured with OAuth2
    Then calling the endpoint without a token should result in status beetween 400 and 403
    And calling the endpoint with an invalid token should result in status beetween 400 and 403
    And calling the endpoint with a valid token should result in status beetween 200 and 299
