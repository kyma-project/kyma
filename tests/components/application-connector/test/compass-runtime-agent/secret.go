package compass_runtime_agent

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"reflect"
	"testing"
)

var secretRepos map[string]v1.SecretInterface

type secretData struct {
	password     string
	username     string
	crt          string
	key          string
	clientId     string
	clientSecret string
}

type SecretClient interface {
	Compare(actual, expected string) bool
	Get() (secretData, error)
}

type Secret struct {
	t          *testing.T
	secret     v1.SecretInterface
	secretName string
}

func NewClient(secret v1.SecretInterface, secretName string) SecretClient {
	return &Secret{
		secret:     secret,
		secretName: secretName,
	}
}

func newSecretsInterface(namespace string) (v1.SecretInterface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Warnf("Failed to read in cluster config: %s", err.Error())
		logrus.Info("Trying to initialize with local config")
		return nil, errors.Wrap(err, "Failed to read in cluster config")
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset.CoreV1().Secrets(namespace), nil
}

func (s Secret) Get() (secretData, error) {
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

func (s Secret) Compare(actual, expected string) bool {

	if len(secretRepos) != 2 {
		secretRepos = make(map[string]v1.SecretInterface)

		expectedSecretRepo, _ := newSecretsInterface("test") //TODO hardcoded, tried os.Getenv() didn't work
		actualSecretRepo, _ := newSecretsInterface(actualSecretNamespace)
		secretRepos["expectedSecretRepo"] = expectedSecretRepo
		secretRepos["actualSecretRepo"] = actualSecretRepo

	}

	secretExpectedClient := NewClient(secretRepos["expectedSecretRepo"], expected)
	secretActualClient := NewClient(secretRepos["actualSecretRepo"], actual)

	actualSecret, err := secretActualClient.Get()
	if err != nil {
		s.t.Log(err, "Failed to get actual secret from cluster")
		errors.Wrap(err, "Failed to get actual secret from cluster")
		return false
	}

	expectedSecret, err := secretExpectedClient.Get()
	if err != nil {
		s.t.Log(err, "Failed to get expected secret from cluster")
		errors.Wrap(err, "Failed to get expected secret from cluster")
		return false
	}

	return reflect.DeepEqual(actualSecret, expectedSecret)
}
