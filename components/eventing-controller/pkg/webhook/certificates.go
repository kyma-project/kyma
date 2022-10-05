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

type Certificates struct {
	secretHandler ISecretHandler
	crdName       string
	secretName    string
	client.Client
	ctx    context.Context
	logger *logr.Logger
}

type ISecretHandler interface {
	ensureSecret(secretName, secretNamespace, serviceName string) (*corev1.Secret, error)
}

type SecretHandler struct {
	certHandler ICertificateHandler
	client.Client
	ctx    context.Context
	logger *logr.Logger
}

type ICertificateHandler interface {
	buildCert(namespace, serviceName string) (cert []byte, key []byte, err error)
	isValidCertificate(cert, key []byte) (bool, error)
}

type CertificateHandler struct {
	clientGoCert IClientGoCert
	ctx          context.Context
	logger       *logr.Logger
}

func NewCertificates(ctx context.Context,
	client client.Client,
	logger *logr.Logger,
	crdName string,
	secretName string) *Certificates {
	return &Certificates{
		secretHandler: &SecretHandler{
			certHandler: &CertificateHandler{
				clientGoCert: &ClientGoCert{},
				ctx:          ctx,
				logger:       logger,
			},
			Client: client,
			logger: logger,
			ctx:    ctx,
		},
		ctx:        ctx,
		Client:     client,
		logger:     logger,
		crdName:    crdName,
		secretName: secretName,
	}
}

func (r *Certificates) Setup() error {
	if err := apiextensionsv1.AddToScheme(r.Client.Scheme()); err != nil {
		return xerrors.Errorf("while adding apiextensions.v1 schema to k8s client: %w", err)
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.crdName}, crd)
	if err != nil {
		return xerrors.Errorf("failed to get %s crd: %w", r.crdName, err)
	}

	// TODO:
	if contains, msg := r.containsConversionWebhookClientConfig(crd); !contains {
		return xerrors.Errorf("failed to validate CRD webhook: %s", msg)
	}

	webhookServiceNamespace := crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace
	webhookServiceName := crd.Spec.Conversion.Webhook.ClientConfig.Service.Name

	return r.ensureWebhookCertificate(r.secretName, webhookServiceNamespace, webhookServiceName)
}

func (r *Certificates) addCertToConversionWebhook(caBundle []byte) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := r.Client.Get(r.ctx, types.NamespacedName{Name: r.crdName}, crd)
	if err != nil {
		return xerrors.Errorf("failed to get APIRule crd: %w", err)
	}

	if contains, msg := r.containsConversionWebhookClientConfig(crd); !contains {
		return errors.Errorf("while validating CRD to be CaBundle injectable,: %s", msg)
	}

	crd.Spec.Conversion.Webhook.ClientConfig.CABundle = caBundle
	err = r.Client.Update(r.ctx, crd)
	if err != nil {
		return xerrors.Errorf("while updating CRD with Conversion webhook caBundle: %w", err)
	}
	return nil
}

func (r *Certificates) ensureWebhookCertificate(secretName, secretNamespace, serviceName string) error {
	r.logger.Info("ensuring webhook secret")

	secret, err := r.secretHandler.ensureSecret(secretName, secretNamespace, serviceName)
	if err != nil {
		return err
	}

	caBundle := secret.Data[CertName]
	if err = r.addCertToConversionWebhook(caBundle); err != nil {
		return xerrors.Errorf("couldn't inject webhook caBundle: %w", err)
	}

	return nil
}

func (r *Certificates) containsConversionWebhookClientConfig(
	crd *apiextensionsv1.CustomResourceDefinition) (bool, string) {
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

func (r *SecretHandler) ensureSecret(secretName, secretNamespace, serviceName string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	r.logger.Info("ensuring webhook secret")

	err := r.Client.Get(r.ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil && !apiErrors.IsNotFound(err) {
		return nil, xerrors.Errorf("failed to get webhook secret: %w", err)
	}

	if apiErrors.IsNotFound(err) {
		r.logger.Info("creating webhook secret")
		if secret, err = r.createSecret(secretName, secretNamespace, serviceName); err != nil {
			return nil, err
		}
		return secret, nil
	}

	_, err = r.isValidSecret(secret)
	if err != nil {
		r.logger.Error(err, "invalid secret")
		if err = r.updateSecret(secret, serviceName); err != nil {
			return secret, err
		}
	}

	return secret, nil
}

func (r *SecretHandler) createSecret(name, namespace, serviceName string) (*corev1.Secret, error) {
	certificate, key, err := r.certHandler.buildCert(namespace, serviceName)
	if err != nil {
		return nil, xerrors.Errorf("failed to build cert: %w", err)
	}

	secret := r.buildSecret(name, namespace, certificate, key)

	if err = r.Client.Create(r.ctx, secret); err != nil {
		return nil, xerrors.Errorf("failed to create secret: %w", err)
	}

	return secret, nil
}

func (r *SecretHandler) buildSecret(name, namespace string, cert []byte, key []byte) *corev1.Secret {
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

func (r *SecretHandler) updateSecret(secret *corev1.Secret, serviceName string) error {
	r.logger.Info("updating pre-exiting secret")

	certificate, key, err := r.certHandler.buildCert(secret.Namespace, serviceName)
	if err != nil {
		return err
	}

	newSecret := r.buildSecret(secret.Name, secret.Namespace, certificate, key)

	secret.Data = newSecret.Data
	if err = r.Client.Update(r.ctx, secret); err != nil {
		return xerrors.Errorf("failed to update secret: %w", err)
	}

	return nil
}

func (r *SecretHandler) isValidSecret(s *corev1.Secret) (bool, error) {
	if !r.hasRequiredKeys(s.Data) {
		return false, xerrors.New("secret data value is missing")
	}
	return r.certHandler.isValidCertificate(s.Data[CertName], s.Data[KeyName])
}

func (r *CertificateHandler) isValidCertificate(cert, key []byte) (bool, error) {
	if err := r.verifyCertificate(cert); err != nil {
		return false, err
	}
	if err := r.verifyKey(key); err != nil {
		return false, err
	}
	return true, nil
}

func (r *SecretHandler) hasRequiredKeys(data map[string][]byte) bool {
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

func (r *CertificateHandler) buildCert(namespace, serviceName string) ([]byte, []byte, error) {
	r.logger.Info("creating certificate")

	crt, key, err := r.generateCertificates(serviceName, namespace)
	if err != nil {
		return nil, nil, xerrors.Errorf("failed to generate certificates: %w", err)
	}

	return crt, key, nil
}

func (r *CertificateHandler) generateCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	altNames := r.serviceAltNames(serviceName, namespace)
	return r.clientGoCert.generateSelfSignedCertKey(altNames[0], nil, altNames)
}

func (r *CertificateHandler) serviceAltNames(serviceName, namespace string) []string {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)

	return []string{
		serviceName,
		namespacedServiceName,
		commonName,
		serviceHostname,
	}
}

func (r *CertificateHandler) verifyCertificate(c []byte) error {
	certificate, err := cert.ParseCertsPEM(c)
	if err != nil {
		return xerrors.Errorf("failed to parse certificate data: %w", err)
	}
	// certificate is self signed. So we use it as a root cert
	root, err := cert.NewPoolFromBytes(c)
	if err != nil {
		return xerrors.Errorf("failed to parse root certificate data: %w", err)
	}
	// make sure the certificate is valid for the next 10 days. Otherwise it will be recreated.
	_, err = certificate[0].Verify(x509.VerifyOptions{CurrentTime: time.Now().Add(10 * 24 * time.Hour), Roots: root})
	if err != nil {
		return xerrors.Errorf("certificate verification failed: %w", err)
	}
	return nil
}

func (r *CertificateHandler) verifyKey(k []byte) error {
	b, _ := pem.Decode(k)
	key, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return xerrors.Errorf("failed to parse key data: %w", err)
	}
	if err = key.Validate(); err != nil {
		return xerrors.Errorf("key verification failed: %w", err)
	}
	return nil
}
