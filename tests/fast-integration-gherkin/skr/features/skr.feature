Feature: SKR test
    These scenarios cover SAP Kyma runtime

    Background: Provision SKR
        Given SKR is provisioned

    Scenario: Function should not be reachable when commerce backend is up without an authorization token
        Given Commerce Backend is set up
        When Function is called without an authorization token
        Then The function returns an error