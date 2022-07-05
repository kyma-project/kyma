Feature: Exposing unsecure API and then securing it with OAuth2 and JWT

  Scenario: Securing an unsecured API with OAuth2 and calling it without a token
    Given There is an unsecured API with all paths available without authorization
    When Endpoint is secured with OAuth2
    Then Calling the endpoint without a token should result in status beetween 400 and 403
    And Calling the endpoint with a invalid token should result in status beetween 400 and 403
    And Calling the endpoint with a valid token should result in status beetween 200 and 299
