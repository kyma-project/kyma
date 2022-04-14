Feature: Ingress Certificate 

    Checking the validity of the Ingress certificate

    Background: Fetch the Ingress certificate
        Given The "kyma-gateway-certs" secret is retrieved from "istio-system" namespace

    Scenario: Check the existence of the Ingress certificate
        Then Ingress certificate data should not be empty

    Scenario: Check the validity of the Ingress certificate
        Given The certificate is extracted from the secret data
        When The date of today is set
        Then The validity date of the certificate should be after the date of today
        And The validity date of the certificate should not be earlier than the date of today