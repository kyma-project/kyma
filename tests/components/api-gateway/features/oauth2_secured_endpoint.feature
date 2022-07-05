Feature: Calling secured OAuth2 API endpoint with different tokens

  Scenario: Calling an secured API without a token
    Given There is an endpoint secured with OAuth2 introspection
    Then Calling the endpoint without a token should result in status between 400 and 403
    And Calling the endpoint with a invalid token should result in status between 400 and 403
    And Calling the endpoint with a valid token should result in status between 200 and 299