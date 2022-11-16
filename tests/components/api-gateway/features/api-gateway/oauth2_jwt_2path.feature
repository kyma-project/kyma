Feature: Exposing different endpoints on one deployment with different authorization strategies (OAuth2 and JWT)

  Scenario: OAuth2JWTTwoPaths: Calling the secured OAuth2 endpoint without a token
    Given OAuth2JWTTwoPaths: There is a deployment secured with OAuth2 on path /headers and JWT on path /image
    Then OAuth2JWTTwoPaths: Calling the "/headers" endpoint without a token should result in status between 400 and 403
    And OAuth2JWTTwoPaths: Calling the "/image" endpoint without a token should result in status between 400 and 403
    And OAuth2JWTTwoPaths: Calling the "/headers" endpoint with an invalid token should result in status between 400 and 403
    And OAuth2JWTTwoPaths: Calling the "/image" endpoint with an invalid token should result in status between 400 and 403
    And OAuth2JWTTwoPaths: Calling the "/headers" endpoint with a valid "OAuth2" token should result in status between 200 and 299
    And OAuth2JWTTwoPaths: Calling the "/image" endpoint with a valid "JWT" token should result in status between 200 and 299
    #TODO: Cross calling should not be allowed
