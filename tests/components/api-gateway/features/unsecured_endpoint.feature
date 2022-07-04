Feature: Calling unsecured API endpoint

  Scenario: Calling an unsecured API without token should be succesfull
    Given There is an unsecured endpoint
    Then Calling the endpoint without a token should result in status beetween 200 and 299
  Scenario: Calling an unsecured API with any token should be succesfull
    Given There is an unsecured endpoint
    Then Calling the endpoint with any token should result in status beetween 200 and 299