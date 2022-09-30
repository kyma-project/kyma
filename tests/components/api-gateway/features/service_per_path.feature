Feature: Exposing two services on path level

  Scenario: Service per path
    Given Service per path: There are two endpoints exposed with different services
    Then Service per path: Calling the endpoint "/headers" and "/hello" with any token should result in status between 200 and 299
    And Service per path: Calling the endpoint "/headers" and "/hello" without a token should result in status between 200 and 299

  Scenario: Service override
    Given Multiple service with overrides: There are two endpoints exposed with different services, one on spec level and one on rule level
    Then Multiple service with overrides: Calling the endpoint "/headers" and "/hello" with any token should result in status between 200 and 299
    And Multiple service with overrides: Calling the endpoint "/headers" and "/hello" without a token should result in status between 200 and 299
