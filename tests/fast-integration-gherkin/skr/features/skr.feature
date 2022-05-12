Feature: SKR test
    These scenarios cover SAP Kyma runtime

    Background: Provision SKR
        Given SKR is provisioned

    Scenario: Initial OIDC config is applied on shoot cluster
        Then "Initial" OIDC config is applied on the shoot cluster

    Scenario: Initial OIDC config is part of the kubeconfig
        Then "Initial" OIDC config is part of the kubeconfig

    Scenario: Admin binding exists for the target user
        Then Admin binding exists for "old" user

    Scenario: SKR Service can be updated with OIDC config
        When SKR service is updated
        Then The operation response should have a succeeded state

    Scenario: Get runtime status after SKR service is updated with OIDC config
        When SKR service is updated
        Then Runtime Status should be fetched successfully

    Scenario: Updated OIDC config is applied on shoot cluster
        When SKR service is updated
        Then "Updated" OIDC config is applied on the shoot cluster

    Scenario: Updated OIDC config is part of the kubeconfig
        When SKR service is updated
        Then "Updated" OIDC config is part of the kubeconfig

    Scenario: Admin binding exists for the target user after updating the SKR service
        When SKR service is updated
        Then Admin binding exists for "old" user

    Scenario: SKR Service Admins can be updated
        When The admins for the SKR service are updated
        Then The operation response should have a succeeded state

    Scenario: Get runtime status after SKR service instance admins are updated
        When The admins for the SKR service are updated
        Then Runtime status should be fetched successfully

    Scenario: New cluster admins are configured correctly after SKR service admins are updated
        When The admins for the SKR service are updated
        Then Admin binding exists for "new" user

    Scenario: Old cluster admin no longer exists after SKR service admins are updated
        When The admins for the SKR service are updated
        Then The old admin no longer exists for the SKR service instance

    Scenario: Function should be reachable when commerce backend is up using a correct authorization token
        Given Commerce Backend is set up
        When Function is called using a correct authorization token
        Then The function should be reachable

    Scenario: Function should not be reachable when commerce backend is up without an authorization token
        Given Commerce Backend is set up
        When Function is called without an authorization token
        Then The function returns an error

    Scenario: order.created.v1 legacy event should trigger the lastorder function
        Given Commerce Backend is set up
        When A legacy event is sent
        Then The event should be received correctly