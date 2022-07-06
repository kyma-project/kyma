Feature: Exposing unsecure API and then securing it with OAuth2

  Scenario: UnsecureToSecureOAuth2: Securing an unsecured API with OAuth2
    Given UnsecureToSecureOAuth2: There is an unsecured API with all paths available without authorization
    And UnsecureToSecureOAuth2: The endpoint is reachable
    When UnsecureToSecureOAuth2: Endpoint is secured with OAuth2
    Then UnsecureToSecureOAuth2: Calling the endpoint without a token should result in status beetween 400 and 403
    And UnsecureToSecureOAuth2: Calling the endpoint with a invalid token should result in status beetween 400 and 403
    And UnsecureToSecureOAuth2: Calling the endpoint with a valid token should result in status beetween 200 and 299
