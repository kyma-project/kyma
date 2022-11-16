package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	"os"
	"path"
)

var certDir string
var serviceName string
var serviceNamespace string
var webhookName string
var logFormat string
var logLevel string

func main() {
	flag.StringVar(&certDir, "cert-dir", "", "Path to server certificate directory")
	flag.StringVar(&serviceName, "service-name", "", "Common name of service")
	flag.StringVar(&serviceNamespace, "service-namespace", "", "Namespace of service")
	flag.StringVar(&webhookName, "webhook-name", "", "Name of webhook config to set CA")
	flag.StringVar(&logFormat, "log-format", "text", "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error, fatal)")
	flag.Parse()

	if err := validateFlags(); err != nil {
		panic(err.Error())
	}

	log := func() *zap.Logger {
		log, err := logger.New(logger.JSON, logger.INFO)
		if err != nil {
			panic(err.Error())
		}
		if err = logger.InitKlog(log, logger.INFO); err != nil {
			panic(err.Error())
		}

		return log.WithContext().Desugar()
	}()

	if err := run(log); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func run(log *zap.Logger) error {
	ctx := context.Background()

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	certificate, key, err := generateCert(serviceName, serviceNamespace)

	err = os.MkdirAll(certDir, 0777)
	if err != nil {
		return fmt.Errorf("failed to create cert dir: %v", err)
	}

	err = os.WriteFile(path.Join(certDir, "tls.crt"), certificate, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write tls.crt: %v", err)
	}

	err = os.WriteFile(path.Join(certDir, "tls.key"), key, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write tls.key: %v", err)
	}

	webhookConfig, err := clientset.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Get(ctx, webhookName, metav1.GetOptions{})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return fmt.Errorf("webhook %s not found: %v", webhookName, err)
		}
		return err
	}

	for i := range webhookConfig.Webhooks {
		webhookConfig.Webhooks[i].ClientConfig.CABundle = certificate
	}

	updatedConfig, err := clientset.AdmissionregistrationV1().
		ValidatingWebhookConfigurations().
		Update(ctx, webhookConfig, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update webhook configuration: %v", err)
	}

	for i := range updatedConfig.Webhooks {
		log.Info(fmt.Sprintf("updated webhook config: %s, with caBundle bytes total: %v",
			updatedConfig.Name,
			len(updatedConfig.Webhooks[i].ClientConfig.CABundle)))
	}

	return nil
}

func generateCert(serviceName, namespace string) ([]byte, []byte, error) {
	cn := fmt.Sprintf("%s.%s.svc", serviceName, namespace)
	names := []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.cluster.local", cn),
	}
	return cert.GenerateSelfSignedCertKey(cn, nil, names)
}

func validateFlags() error {
	if certDir == "" {
		return errors.New("--cert-dir is required")
	}
	if serviceName == "" {
		return errors.New("--service-name is required")
	}
	if serviceNamespace == "" {
		return errors.New("--service-namespace is required")
	}
	if webhookName == "" {
		return errors.New("--webhook-name is required")
	}
	return nil
}
