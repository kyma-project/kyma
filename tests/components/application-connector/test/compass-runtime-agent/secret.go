package compass_runtime_agent

import (
	"context"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type secretData struct {
	password     string
	username     string
	crt          string
	key          string
	clientId     string
	clientSecret string
}

//go:generate mockery --name=SecretClient
type SecretClient interface {
	Compare(actual, expected string, cli kubernetes.Interface) bool
	GetSecretData() (secretData, error)
}

type Secret struct {
	secret     v1.SecretInterface
	secretName string
}

func NewClient(secret v1.SecretInterface, secretName string) SecretClient {
	return &Secret{
		secret:     secret,
		secretName: secretName,
	}
}

func (s Secret) GetSecretData() (secretData, error) {
	secret, err := s.secret.Get(context.Background(), s.secretName, metav1.GetOptions{})

	if err != nil {
		return secretData{}, errors.Wrap(err, "Failed to get secret from cluster")
	}

	return secretData{
		username:     string(secret.Data["username"]),
		password:     string(secret.Data["password"]),
		crt:          string(secret.Data["crt"]),
		key:          string(secret.Data["key"]),
		clientId:     string(secret.Data["clientId"]),
		clientSecret: string(secret.Data["clientSecret"]),
	}, nil
}

func (s Secret) Compare(actual, expected string, cli kubernetes.Interface) bool {

	expectedSecretRepo := cli.CoreV1().Secrets("test")
	actualSecretRepo := cli.CoreV1().Secrets(actualSecretNamespace)

	secretExpectedClient := NewClient(expectedSecretRepo, expected)
	secretActualClient := NewClient(actualSecretRepo, actual)

	actualSecret, err := secretActualClient.GetSecretData()
	if err != nil {
		//s.t.Log(err, "Failed to get actual secret from cluster")
		errors.Wrap(err, "Failed to get actual secret from cluster")
		return false
	}

	expectedSecret, err := secretExpectedClient.GetSecretData()
	if err != nil {
		//s.t.Log(err, "Failed to get expected secret from cluster")
		errors.Wrap(err, "Failed to get expected secret from cluster")
		return false
	}

	secretsEqual := actualSecret == expectedSecret

	if !secretsEqual {
		//s.t.Log("Secrets aren't equal")
		errors.New("Secrets aren't equal")
		return false
	}

	return secretsEqual
}
