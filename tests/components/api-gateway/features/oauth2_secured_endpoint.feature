Feature: Calling secured OAuth2 API endpoint with different tokens

  Scenario: OAuth2: Exposing an endpoint with OAuth2
    Given OAuth2: There is an endpoint secured with OAuth2 introspection
    Then OAuth2: Calling the endpoint without a token should result in status between 400 and 403
    And OAuth2: Calling the endpoint with a invalid token should result in status between 400 and 403
    And OAuth2: Calling the endpoint with a valid token should result in status between 200 and 299
