Feature: Exposing two services on path level

  Scenario: Unsecured: Calling an unsecured API endpoint
    Given Service per path: There are two endpoints exposed with different services
    Then Service per path: Calling the endpoint "/headers" and "/hello" with any token should result in status between 200 and 299
    And Service per path: Calling the endpoint "/headers" and "/hello" without a token should result in status between 200 and 299
