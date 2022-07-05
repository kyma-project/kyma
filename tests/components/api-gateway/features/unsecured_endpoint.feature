Feature: Calling unsecured API endpoint

  Scenario: Calling an unsecured endpoint without token should be succesfull
    Given There is an unsecured endpoint
    Then Calling the endpoint without a token should result in status between 200 and 299
    And Calling the endpoint with any token should result in status between 200 and 299