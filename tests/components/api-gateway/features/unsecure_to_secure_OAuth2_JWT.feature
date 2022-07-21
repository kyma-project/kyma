Feature: Exposing unsecure API and then securing it with JWT and OAuth2 on two paths

  Scenario: UnsecureToSecureOAuth2JWT: Securing an unsecured API with OAuth2 and JWT
    Given UnsecureToSecureOAuth2JWT: There is an unsecured API with all paths available without authorization
    And UnsecureToSecureOAuth2JWT: The endpoint is reachable
    When UnsecureToSecureOAuth2JWT: API is secured with OAuth2 on path /headers and JWT on path /image
    Then UnsecureToSecureOAuth2JWT: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And UnsecureToSecureOAuth2JWT: Calling the "/image" endpoint without a token should result in status between 400 and 403
    And UnsecureToSecureOAuth2JWT: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And UnsecureToSecureOAuth2JWT: Calling the "/image" endpoint with an invalid token should result in status between 400 and 403
    And UnsecureToSecureOAuth2JWT: Calling the "/headers" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    And UnsecureToSecureOAuth2JWT: Calling the "/image" endpoint with a valid "JWT" token should result in status between 200 and 299
