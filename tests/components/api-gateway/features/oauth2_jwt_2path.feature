Feature: Exposing different endpoints on one deployment with different authorization strategies (OAuth2 and JWT)

  Scenario: Calling the secured OAuth2 endpoint without a token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "OAuth2" endpoint "without" a token should result in status beetween 400 and 403
  Scenario: Calling the secured JWT endpoint without a token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "JWT" endpoint "without" a token should result in status beetween 400 and 403
  Scenario: Calling the secured OAuth2 endpoint with an invalid token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "OAuth2" endpoint with an "invalid" token should result in status beetween 400 and 403
  Scenario: Calling the secured JWT endpoint with an invalid token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "JWT" endpoint with an "invalid" token should result in status beetween 400 and 403
  #TODO: Cross calling should not be allowed
  Scenario: Calling the secured OAuth2 endpoint with a valid token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "OAuth2" endpoint with a "valid" token should result in status beetween 200 and 299
  Scenario: Calling the secured JWT endpoint with a valid token
    Given There are endpoints secured with OAuth2 and JWT
    Then Calling the "JWT" endpoint with a "valid" token should result in status beetween 200 and 299
