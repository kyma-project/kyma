package azurevault

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type azureKeyvaultInterface interface {
	GetSecret(secretName string) (keyvault.SecretBundle, error)
	GetCertificate(certName string) (keyvault.CertificateBundle, error)
	GetKey(keyName string) (keyvault.KeyBundle, error)
}

type azureKeyvault struct {
	client   keyvault.BaseClient
	vaultURL string
}

func getAzureKeyvaultClient(azureClientID string, azureClientSecret string, vaultURL string) *azureKeyvault {
	return &azureKeyvault{
		client:   getClient(azureClientID, azureClientSecret),
		vaultURL: vaultURL,
	}
}

func getClient(azureClientID string, azureClientSecret string) keyvault.BaseClient {
	client := keyvault.New()
	client.Sender = autorest.CreateSender()

	client.Authorizer = autorest.NewBearerAuthorizerCallback(client.Sender, func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		env := azure.PublicCloud
		keyVaultOauthConfig, err := getAzureOAuthConfig(env.ActiveDirectoryEndpoint, tenantID)
		if err != nil {
			return nil, err
		}

		keyVaultSpt, err := adal.NewServicePrincipalToken(*keyVaultOauthConfig, azureClientID, azureClientSecret, resource)
		if err != nil {
			return nil, err
		}

		return autorest.NewBearerAuthorizer(keyVaultSpt), nil
	})

	return client
}

func getAzureOAuthConfig(endpoint, tenantID string) (*adal.OAuthConfig, error) {
	oauthConfig, err := adal.NewOAuthConfig(endpoint, tenantID)
	if err != nil {
		return nil, err
	}

	if oauthConfig == nil {
		return nil, fmt.Errorf("Unable to configure OAuthConfig for tenant %s", tenantID)
	}

	return oauthConfig, nil
}

//GetSecret .
func (ac *azureKeyvault) GetSecret(secretName string) (keyvault.SecretBundle, error) {
	ctx := context.Background()
	return ac.client.GetSecret(ctx, ac.vaultURL, secretName, "")
}

//GetCertificate .
func (ac *azureKeyvault) GetCertificate(certName string) (keyvault.CertificateBundle, error) {
	ctx := context.Background()
	return ac.client.GetCertificate(ctx, ac.vaultURL, certName, "")
}

//GetKey .
func (ac *azureKeyvault) GetKey(keyName string) (keyvault.KeyBundle, error) {
	ctx := context.Background()
	return ac.client.GetKey(ctx, ac.vaultURL, keyName, "")
}
