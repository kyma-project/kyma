Feature: SKR test
    These scenarios cover SAP Kyma runtime

    Background: Provision SKR
        Given SKR is provisioned

    Scenario: Initial OIDC config is applied on shoot cluster
        Then "Initial" OIDC config is applied on the shoot cluster
