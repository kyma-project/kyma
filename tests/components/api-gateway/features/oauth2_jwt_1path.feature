Feature: Exposing one endpoint with different authorization strategies (OAuth2 and JWT)

  Scenario: OAuth2JWT1Path: Calling an endpoint secured with OAuth2 and JWT with a valid OAuth2 token
    Given OAuth2JWT1Path: There is an deployment secured with both JWT and OAuth2 introspection on path /image
    Then OAuth2JWT1Path: Calling the "/image" endpoint without a token should result in status between 400 and 403
    And OAuth2JWT1Path: Calling the "/image" endpoint with a invalid token should result in status between 400 and 403
    And OAuth2JWT1Path: Calling the "/image" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    And OAuth2JWT1Path: Calling the "/image" endpoint with a valid "JWT" token should result in status between 200 and 299
