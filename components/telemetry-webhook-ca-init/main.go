package main

import (
	"bytes"
	"context"
	"flag"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

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

	err = ensureCaSecret(ctx, clientset, secretName, secretNamespace, caName, log)
	if err != nil {
		panic(err.Error())
	}

	//caBundle, err := internal.CreateCABundle(caName)
	//if err != nil {
	//	log.WithTracing(ctx).Error(err, "failed to create CA bundle")
	//	os.Exit(1)
	//}
	//
	//err = os.MkdirAll(certDir, 0777)
	//if err != nil {
	//	log.WithTracing(ctx).Error(err, "failed to create certs directory")
	//	os.Exit(1)
	//}
	//
	//err = writeFile(certDir+"tls.crt", caBundle.ServerCert)
	//if err != nil {
	//	log.WithTracing(ctx).Error(err, "failed to write tls.crt")
	//	os.Exit(1)
	//}
	//
	//err = writeFile(certDir+"tls.key", caBundle.ServerPrivKey)
	//if err != nil {
	//	log.WithTracing(ctx).Error(err, "failed to write tls.key")
	//	os.Exit(1)
	//}

	// TODO get webhook config and set caBundle
}

func ensureCaSecret(ctx context.Context, clientset *kubernetes.Clientset, secretName, secretNamespace, serviceName string, log *zap.SugaredLogger) error {
	log.Info("ensuring ca secret")
	secret, err := clientset.CoreV1().Secrets(secretName).Get(ctx, secretNamespace, v1.GetOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to get ca secret")
	}

	if apiErrors.IsNotFound(err) {
		log.Info("creating ca secret")
		return createSecret(ctx, client, secretName, secretNamespace, serviceName)
	}

	log.Info("updating pre-exiting webhook secret")
	if err := updateSecret(ctx, client, log, secret, serviceName); err != nil {
		return errors.Wrap(err, "failed to update secret")
	}
	return nil
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

func buildSecret(name, namespace, serviceName string) (*corev1.Secret, error) {
	cert, key, err := generateWebhookCertificates(serviceName, namespace)
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

func writeFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.Write(sCert.Bytes())
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
