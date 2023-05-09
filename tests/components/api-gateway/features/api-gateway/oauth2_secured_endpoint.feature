Feature: Exposing an endpoint with OAuth2

  Scenario: OAuth2: Exposing an endpoint with OAuth2
    Given OAuth2: There is an endpoint /headers secured with OAuth2 introspection and requiring scope read
    And OAuth2: There is an endpoint /ip secured with OAuth2 introspection and requiring scope special-scope
    Then OAuth2: Calling the "/headers" endpoint without a token should result in status between 401 and 401
    And OAuth2: Calling the "/headers" endpoint with a invalid token should result in status between 401 and 401
    And OAuth2: Calling the "/headers" endpoint with a valid token with scope claim read should result in status between 200 and 299
    And OAuth2: Calling the "/get" endpoint with a valid token with scope claim read should result in status between 403 and 403
