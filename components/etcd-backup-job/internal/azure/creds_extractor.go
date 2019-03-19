package azure

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreCli "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	secretKeyAccountName = "storage-account"
	secretKeyAccountKey  = "storage-key"
)

// ExtractCredsFromSecretOutput contains all information which will be returned from ExtractCredsFromSecret function
type ExtractCredsFromSecretOutput struct {
	AccountName string
	AccountKey  string
}

// ExtractCredsFromSecret returns account info for ABS from given k8s Secret
func ExtractCredsFromSecret(absSecretName string, secretCli coreCli.SecretInterface) (*ExtractCredsFromSecretOutput, error) {
	secret, err := secretCli.Get(absSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting secret %q", absSecretName)
	}

	if secret.Data == nil || len(secret.Data) == 0 {
		return nil, errors.Errorf("Secret %q data cannot be empty", absSecretName)
	}

	accountName, accountKey := secret.Data[secretKeyAccountName], secret.Data[secretKeyAccountKey]

	return &ExtractCredsFromSecretOutput{
		AccountKey:  string(accountKey),
		AccountName: string(accountName),
	}, nil
}
