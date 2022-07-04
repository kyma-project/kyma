Feature: Exposing unsecure API and then securing it with JWT

  Scenario: Securing an unsecured API with JWT and calling it without a token
    Given There is an unsecured API with all paths available without authorization
    When Endpoint is secured with JWT
    Then Calling the endpoint without a token should result in status beetween 400 and 403
  Scenario: Securing an unsecured API with JWT and calling it with a invalid token
    Given There is an unsecured API with all paths available without authorization
    When Endpoint is secured with JWT
    Then Calling the endpoint with a "invalid" token should result in status beetween 400 and 403
  Scenario: Securing an unsecured API with JWT and calling it with a valid token
    Given There is an unsecured API with all paths available without authorization
    When Endpoint is secured with JWT
    Then Calling the endpoint with a "valid" token should result in status beetween 200 and 299