package compass_runtime_agent

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type secretData struct {
	password     string
	username     string
	crt          string
	key          string
	clientId     string
	clientSecret string
}

type Client interface {
	GetSecret() (secretData, error)
}

type client struct {
	secretsClient v1.SecretInterface
	secretName    string
}

func NewClient(secrets v1.SecretInterface, secretName string) Client {
	return &client{
		secretsClient: secrets,
		secretName:    secretName,
	}
}

func newSecretsInterface(namespace string) (v1.SecretInterface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Warnf("Failed to read in cluster config: %s", err.Error())
		logrus.Info("Trying to initialize with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset.CoreV1().Secrets(namespace), nil
}

func initalizeSecretInterface(namespace string) (v1.SecretInterface, error) {
	secretsRepo, err := newSecretsInterface(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create secrets interface")
	}
	return secretsRepo, nil
}

func (c *client) getCredentials() (secretData, error) {
	secret, err := c.secretsClient.Get(context.Background(), c.secretName, metav1.GetOptions{})

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

func (c *client) GetSecret() (secretData, error) {

	secret, err := c.getCredentials()
	if err != nil {
		return secretData{}, errors.Wrap(err, "Failed to save the secret from cluster")
	}

	return secret, nil
}

func SecretDataCompare(actual, expected string, t *testing.T) bool {
	expectedSecretRepo, _ := initalizeSecretInterface(os.Getenv("NAMESPACE"))
	expectedSecretClient := NewClient(expectedSecretRepo, expected)

	expectedSecret, err := expectedSecretClient.GetSecret()
	if err != nil {
		t.Log(err, "Failed to get expected secret from cluster")
		errors.Wrap(err, "Failed to get expected secret from cluster")
		return false
	}

	actualSecretRepo, _ := initalizeSecretInterface(actualSecretNamespace)
	actualSecretClient := NewClient(actualSecretRepo, actual)

	actualSecret, err := actualSecretClient.GetSecret()
	if err != nil {
		t.Log(err, "Failed to get actual secret from cluster")
		errors.Wrap(err, "Failed to get actual secret from cluster")
		return false
	}

	fmt.Println(actualSecret)

	return reflect.DeepEqual(actualSecret, expectedSecret)
}
