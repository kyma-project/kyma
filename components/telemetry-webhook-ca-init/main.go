package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	"os"
)

var certDir string
var serviceName string
var serviceNamespace string
var webhookName string

func main() {
	flag.StringVar(&certDir, "cert-dir", "", "Path to server certificate directory")
	flag.StringVar(&serviceName, "service-name", "", "Common name of service")
	flag.StringVar(&serviceNamespace, "service-namespace", "", "Namespace of service")
	flag.StringVar(&webhookName, "webhook-name", "", "Name of webhook config to set CA")
	flag.Parse()

	if err := validateFlags(); err != nil {
		panic(err.Error())
	}
	ctx := context.Background()

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	certificate, key, err := generateCert(serviceName, serviceNamespace)

	err = os.MkdirAll(certDir, 0777)
	if err != nil {
		fmt.Printf("failed to create cert dir: %s", err.Error())
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.crt", certificate)
	if err != nil {
		fmt.Printf("failed to write tls.crt: %s", err.Error())
		os.Exit(1)
	}

	err = writeFile(certDir+"tls.key", key)
	if err != nil {
		fmt.Printf("failed to write tls.key: %s", err.Error())
		os.Exit(1)
	}

	webhookConfig, err := clientset.AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Get(ctx, webhookName, metav1.GetOptions{})

	if apiErrors.IsNotFound(err) {
		fmt.Printf("webhook %s not found", webhookName)
		os.Exit(1)
	} else if err != nil {
		panic(err)
	}

	webhookConfig.Webhooks[0].ClientConfig.CABundle = certificate

	updatedConfig, err := clientset.AdmissionregistrationV1beta1().
		ValidatingWebhookConfigurations().
		Update(ctx, webhookConfig, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("failed to update webhook configuration: %s", err.Error())
		os.Exit(1)
	}

	fmt.Printf("updated webhook config: %s, with caBundle: %v",
		updatedConfig.Name,
		updatedConfig.Webhooks[0].ClientConfig.CABundle)
}

func generateCert(serviceName, namespace string) ([]byte, []byte, error) {
	names := []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
	}
	return cert.GenerateSelfSignedCertKey(names[0], nil, names)
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
