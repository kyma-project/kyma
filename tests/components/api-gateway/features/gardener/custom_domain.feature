Feature: Exposing endpoint with a custom domain
  Scenario: CustomDomain: Register new custom domain on the cluster
    Given CustomDomain: There is an secret with DNS cloud service provider credentials
    Then CustomDomain: Create needed resources

  Scenario: CustomDomain: Calling an unsecured API endpoint with custom domain
    Given CustomDomain: There is an unsecured endpoint
    Then CustomDomain: Calling the endpoint without a token should result in status between 200 and 299
    And CustomDomain: Calling the endpoint with any token should result in status between 200 and 299
 