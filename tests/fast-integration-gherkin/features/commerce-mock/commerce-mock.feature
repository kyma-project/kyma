Feature: Commerce-mock

    Testing Kyma functionality using the commerce mock

    Background: Set up the commerce backend
        Given All pods in the cluster are listed
        And The Commerce backend is set up
        And Loki port is forwarded

    Scenario: Checking structured in-cluster event delivery
        When A structured event is sent
        Then The structured event should be received correctly

    Scenario: Checking binary in-cluster event delivery
        When A binary event is sent
        Then The binary event should be received correctly

    Scenario: Function should be reachable when using a correct authorization token
        When 
        Then

    Scenario: Function should not be reachable when unauthorized
        When 
        Then