Feature: Exposing an unsecured endpoint

  Scenario: Unsecured: Calling an unsecured API endpoint
    Given Unsecured: There is an unsecured endpoint
    Then Unsecured: Calling the endpoint without a token should result in status between 200 and 299
    And Unsecured: Calling the endpoint with any token should result in status between 200 and 299
