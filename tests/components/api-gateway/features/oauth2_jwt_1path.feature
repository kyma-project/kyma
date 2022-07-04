Feature: Exposing one endpoint with different authorization strategies (OAuth2 and JWT)

  Scenario: Calling an endpoint secured with OAuth2 and JWT with a valid OAuth2 token
    Given There is an endpoint secured with both JWT and OAuth2 introspection
    Then Calling the endpoint without a token should result in status beetween 400 and 403
  Scenario: Calling an endpoint secured with OAuth2 and JWT with an invalid token
    Given There is an endpoint secured with both JWT and OAuth2 introspection
    Then Calling it with a "invalid" "OAuth2" token should result in status beetween 200 and 299
  Scenario: Calling an endpoint secured with OAuth2 and JWT with an invalid token
    Given There is an endpoint secured with both JWT and OAuth2 introspection
    Then Calling it with a "invalid" "JWT" token should result in status beetween 200 and 299
  Scenario: Calling an endpoint secured with OAuth2 and JWT with a valid OAuth2 token
    Given There is an endpoint secured with both JWT and OAuth2 introspection
    Then Calling it with a "valid" "OAuth2" token should result in status beetween 200 and 299
  Scenario: Calling an endpoint secured with OAuth2 and JWT with a valid JWT token
    Given There is an endpoint secured with both JWT and OAuth2 introspection
    Then Calling it with a "valid" "JWT" token should result in status beetween 200 and 299
