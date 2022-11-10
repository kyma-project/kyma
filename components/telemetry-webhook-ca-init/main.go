package main

import (
	"context"
	"flag"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"telemetry-webhook-ca-init/internal"

	corev1 "k8s.io/api/core/v1"
	"os"
)

const caCert = "ca-cert"
const caKey = "ca-key"
const secretName = "webhook-ca"
const secretNamespace = "kyma-system"
const caName = "telemetry-validating-webhook-ca"
const retriesOnFailure = 5

// const certDir = "/var/run/telemetry-webhook/"
var certDir string

func main() {
	flag.StringVar(&certDir, "cert-dir", "", "Path to certificate bundle directory")
	flag.Parse()

	// TODO debug
	certDir = "./bin"

	if err := validateFlags(); err != nil {
		panic(err.Error())
	}
	ctx := context.Background()
	logger, err := logger.New("text", "info")
	if err != nil {
		panic(err.Error())
	}
	log := logger.WithContext()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	caSecret, err := getOrCreateCaSecret(ctx, clientset, secretName, secretNamespace, caName, log)
	if err != nil {
		log.Errorf("failed to ensure ca secret: %s", err.Error())
		panic(err.Error())
	}

	ca, found := caSecret.Data[caCert]
	if !found {
		log.Error(err, "invalid secret state: ca-cert not found")
		os.Exit(1)
	}

	key, found := caSecret.Data[caKey]
	if !found {
		log.Error(err, "invalid secret state: ca-key not found")
		os.Exit(1)
	}

	caBundle := &internal.CABundle{
		CA:    ca,
		CAKey: key,
	}

	serverCert, err := internal.GenerateServerCertAndKey(caBundle, "", "")
	if err != nil {
		log.Error(err, "failed to generate server cert")
		os.Exit(1)
	}

	err = os.MkdirAll(certDir, 0777)
	if err != nil {
		log.Error(err, "failed to create cert dir")
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.crt", serverCert.Cert)
	if err != nil {
		log.Error(err, "failed to write tls.crt")
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.key", serverCert.Key)
	if err != nil {
		log.Error(err, "failed to write tls.key")
		os.Exit(1)
	}

	webhookConfig, err := clientset.AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Get(ctx, "", metav1.GetOptions{})

	webhookConfig.Webhooks[0].ClientConfig.CABundle = caBundle.CA

	updatedConfig, err := clientset.AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Update(ctx, webhookConfig, metav1.UpdateOptions{})
	if err != nil {
		log.Error(err, "failed to update webhook configuration")
		os.Exit(1)
	}

	log.Infof("updated webhook config: %s, with caBundle: %v",
		updatedConfig.Name,
		updatedConfig.Webhooks[0].ClientConfig.CABundle)
}

func getOrCreateCaSecret(ctx context.Context, clientset *kubernetes.Clientset, name, namespace, service string, log *zap.SugaredLogger) (*v1.Secret, error) {
	log.Info("ensuring ca secret")
	secret, err := clientset.CoreV1().Secrets(name).Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return secret, nil
	}
	if apiErrors.IsNotFound(err) {
		log.Info("creating ca secret")
		secret, err = buildSecret(name, namespace, service)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create secret object")
		}
		secret, err = clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create secret")
		}
		return secret, nil
	}
	return nil, errors.Wrap(err, "failed to get or create ca secret")
}

func buildSecret(name, namespace, service string) (*corev1.Secret, error) {
	ca, err := internal.GenerateCACert()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate ca certificate")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			caCert: ca.CA,
			caKey:  ca.CAKey,
		},
	}, nil
}

func writeFile(filepath string, data []byte) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func validateFlags() error {
	if certDir == "" {
		return errors.New("--cert-dir flag is required")
	}
	return nil
}
