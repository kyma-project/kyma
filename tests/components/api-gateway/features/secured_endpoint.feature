Feature: Calling secured OAuth2 API endpoint with different tokens

  Scenario: Calling an secured API without a token
    Given There is an endpoint secured with OAuth2 introspection
    Then Calling the endpoint without a token should result in status beetween 400 and 403
  Scenario: Calling an secured API with an invalid token should not be allowed
    Given There is an endpoint secured with OAuth2 introspection
    Then Calling the endpoint with a "invalid" token should result in status beetween 400 and 403
  Scenario: Calling an secured API with an valid token should not be succesfull
    Given There is an endpoint secured with OAuth2 introspection
    Then Calling the endpoint with a "valid" token should result in status beetween 200 and 299