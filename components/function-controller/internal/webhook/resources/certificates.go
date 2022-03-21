package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CertFile       = "server-cert.pem"
	KeyFile        = "server-key.pem"
	DefaultCertDir = "/tmp/k8s-webhook-server/serving-certs"
)

func SetupCertificates(ctx context.Context, secretName, secretNamespace, serviceName string) error {
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}

	return errors.Wrap(EnsureWebhookSecret(ctx, serverClient, secretName, secretNamespace, serviceName), "failed to ensure webhook secret")
}

func serviceAltNames(serviceName, namespace string) []string {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)

	return []string{
		commonName,
		serviceName,
		namespacedServiceName,
		serviceHostname,
	}
}

func GenerateWebhookCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	altNames := serviceAltNames(serviceName, namespace)
	return cert.GenerateSelfSignedCertKey(altNames[0], nil, altNames)
}

func EnsureWebhookSecret(ctx context.Context, client ctrlclient.Client, secretName, secretNamespace, serviceName string) error {
	logger := ctrl.LoggerFrom(ctx)
	secret := &corev1.Secret{}
	logger.Info("ensuring webhook secret")
	err := client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to get webhook secret")
	}

	if apiErrors.IsNotFound(err) {
		logger.Info("creating webhook secret")
		return createSecret(ctx, client, secretName, secretNamespace, serviceName)
	}

	logger.Info("updating pre-exiting webhook secret")
	return errors.Wrap(updateSecret(ctx, client, secret, serviceName), "failed to update secret")
}

func buildSecret(name, namespace, serviceName string) (*corev1.Secret, error) {
	cert, key, err := GenerateWebhookCertificates(serviceName, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate webhook certificates")
	}
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			CertFile: cert,
			KeyFile:  key,
		},
	}, nil
}

func createSecret(ctx context.Context, client ctrlclient.Client, name, namespace, serviceName string) error {
	secret, err := buildSecret(name, namespace, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to create secret object")
	}
	if err := client.Create(ctx, secret); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}
	return nil
}

func updateSecret(ctx context.Context, client ctrlclient.Client, secret *corev1.Secret, serviceName string) error {
	update := false
	if secret.Data != nil {
		for _, key := range []string{CertFile, KeyFile} {
			if _, ok := secret.Data[key]; !ok {
				update = true
				break
			}
		}
	}
	if update || secret.Data == nil {
		newSecret, err := buildSecret(secret.Name, secret.Namespace, serviceName)
		if err != nil {
			return errors.Wrap(err, "failed to create secret object")
		}
		secret.Data = newSecret.Data
		return errors.Wrap(client.Update(ctx, secret), "failed to update secret")
	}
	return nil
}
