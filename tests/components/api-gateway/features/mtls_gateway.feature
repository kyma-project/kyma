Feature: Exposing an unsecured endpoint on mTLS Gateway

  Scenario: mTLSGateway: Calling an unsecured API endpoint
    Given mTLSGateway: There is an unsecured endpoint
    Then mTLSGateway: Calling the endpoint without a token should result in status between 200 and 299
    And mTLSGateway: Calling the endpoint with any token should result in status between 200 and 299
