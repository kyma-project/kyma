package webhook

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"golang.org/x/xerrors"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CertName = "tls.crt"
	KeyName  = "tls.key"
)

type WebhookCertificateHandler struct {
	crdName    string
	secretName string
	client.Client
	ctx    context.Context
	logger *logr.Logger
}

func NewWebhookCertificateHandler(ctx context.Context, client client.Client, logger *logr.Logger, crdName string, secretName string) *WebhookCertificateHandler {
	return &WebhookCertificateHandler{
		ctx:        ctx,
		Client:     client,
		logger:     logger,
		crdName:    crdName,
		secretName: secretName,
	}
}

func (r *WebhookCertificateHandler) SetupCertificates() error {
	if err := apiextensionsv1.AddToScheme(r.Client.Scheme()); err != nil {
		return errors.Wrap(err, "while adding apiextensions.v1 schema to k8s client")
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.crdName}, crd)
	if err != nil {
		return xerrors.Errorf("failed to get %s crd: %w", r.crdName, err)
	}

	if contains, msg := r.containsConversionWebhookClientConfig(crd); !contains {
		return xerrors.Errorf("failed to validate CRD webhook: %s", msg)
	}

	webhookServiceNamespace := crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace
	webhookServiceName := crd.Spec.Conversion.Webhook.ClientConfig.Service.Name

	return r.ensureWebhookCertificate(r.secretName, webhookServiceNamespace, webhookServiceName)
}

func (r *WebhookCertificateHandler) createCABundle(webhookNamespace string, serviceName string) ([]byte, []byte, error) {
	r.logger.Info("creating certificate for webhook")

	certificate, key, err := r.createCert(webhookNamespace, serviceName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to crete cert")
	}
	return certificate, key, nil
}

func (r *WebhookCertificateHandler) addCertToConversionWebhook(caBundle []byte) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.crdName}, crd)
	if err != nil {
		return errors.Wrap(err, "failed to get APIRule crd")
	}

	if contains, msg := r.containsConversionWebhookClientConfig(crd); !contains {
		return errors.Errorf("while validating CRD to be CaBundle injectable,: %s", msg)
	}

	crd.Spec.Conversion.Webhook.ClientConfig.CABundle = caBundle
	err = r.Client.Update(r.ctx, crd)
	if err != nil {
		return errors.Wrap(err, "while updating CRD with Conversion webhook caBundle")
	}
	return nil
}

func (r *WebhookCertificateHandler) ensureWebhookCertificate(secretName, secretNamespace, serviceName string) error {
	secret := &corev1.Secret{}
	r.logger.Info("ensuring webhook secret")

	err := r.Client.Get(r.ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrap(err, "failed to get webhook secret")
	}

	if apiErrors.IsNotFound(err) {
		r.logger.Info("creating webhook secret")
		return r.createSecret(secretName, secretNamespace, serviceName)
	}

	r.logger.Info("updating pre-exiting webhook secret")
	if err := r.updateSecret(secret, serviceName); err != nil {
		return errors.Wrap(err, "failed to update secret")
	}
	return nil
}

func (r *WebhookCertificateHandler) createSecret(name, namespace, serviceName string) error {
	certificate, key, err := r.buildCert(namespace, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to build cert ")
	}

	secret := r.buildSecret(name, namespace, certificate, key)

	if err := r.Client.Create(r.ctx, secret); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	err = r.addCertToConversionWebhook(certificate)
	if err != nil {
		return err
	}
	return nil
}

func (r *WebhookCertificateHandler) containsConversionWebhookClientConfig(crd *apiextensionsv1.CustomResourceDefinition) (bool, string) {
	if crd.Spec.Conversion == nil {
		return false, "conversion not found in " + r.crdName
	}

	if crd.Spec.Conversion.Webhook == nil {
		return false, "conversion webhook not found in " + r.crdName
	}

	if crd.Spec.Conversion.Webhook.ClientConfig == nil {
		return false, "client config for conversion webhook not found in " + r.crdName
	}
	return true, ""
}

func (r *WebhookCertificateHandler) createCert(webhookNamespace string, serviceName string) ([]byte, []byte, error) {

	certificate, key, err := r.buildCert(webhookNamespace, serviceName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build certificate")
	}

	return certificate, key, nil
}

func (r *WebhookCertificateHandler) isValidSecret(s *corev1.Secret) (bool, error) {
	if !r.hasRequiredKeys(s.Data) {
		return false, nil
	}
	if err := r.verifyCertificate(s.Data[CertName]); err != nil {
		return false, err
	}
	if err := r.verifyKey(s.Data[KeyName]); err != nil {
		return false, err
	}

	return true, nil
}

func (r *WebhookCertificateHandler) verifyCertificate(c []byte) error {
	certificate, err := cert.ParseCertsPEM(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate data")
	}
	// certificate is self signed. So we use it as a root cert
	root, err := cert.NewPoolFromBytes(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse root certificate data")
	}
	// make sure the certificate is valid for the next 10 days. Otherwise it will be recreated.
	_, err = certificate[0].Verify(x509.VerifyOptions{CurrentTime: time.Now().Add(10 * 24 * time.Hour), Roots: root})
	if err != nil {
		return errors.Wrap(err, "certificate verification failed")
	}
	return nil
}

func (r *WebhookCertificateHandler) verifyKey(k []byte) error {
	b, _ := pem.Decode(k)
	key, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse key data")
	}
	if err = key.Validate(); err != nil {
		return errors.Wrap(err, "key verification failed")
	}
	return nil
}

func (r *WebhookCertificateHandler) hasRequiredKeys(data map[string][]byte) bool {
	if data == nil {
		return false
	}
	for _, key := range []string{CertName, KeyName} {
		if _, ok := data[key]; !ok {
			return false
		}
	}
	return true
}

func (r *WebhookCertificateHandler) buildCert(namespace, serviceName string) (cert []byte, key []byte, err error) {
	cert, key, err = r.generateWebhookCertificates(serviceName, namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate webhook certificates")
	}

	return cert, key, nil
}

func (r *WebhookCertificateHandler) updateSecret(secret *corev1.Secret, serviceName string) error {
	valid, err := r.isValidSecret(secret)
	if valid {
		return nil
	}
	if err != nil {
		r.logger.Error(err, "invalid certificate")
	}

	certificate, key, err := r.createCABundle(secret.Namespace, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to ensure webhook secret")
	}

	newSecret := r.buildSecret(secret.Name, secret.Namespace, certificate, key)

	secret.Data = newSecret.Data
	if err := r.Client.Update(r.ctx, secret); err != nil {
		return errors.Wrap(err, "failed to update secret")
	}

	if err := r.addCertToConversionWebhook(certificate); err != nil {
		return errors.Wrap(err, "while adding CaBundle to Conversion Webhook for function CRD")
	}
	return nil
}

func (r *WebhookCertificateHandler) generateWebhookCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	altNames := r.serviceAltNames(serviceName, namespace)
	return cert.GenerateSelfSignedCertKey(altNames[0], nil, altNames)
}

func (r *WebhookCertificateHandler) serviceAltNames(serviceName, namespace string) []string {
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

func (r *WebhookCertificateHandler) buildSecret(name, namespace string, cert []byte, key []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			CertName: cert,
			KeyName:  key,
		},
	}
}
